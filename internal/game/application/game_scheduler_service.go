package application

import (
	"GAMERS-BE/internal/game/application/port"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// Redis distributed lock keys
	lockKeyActivation = "scheduler:lock:activation"
	lockKeyDetection  = "scheduler:lock:detection"

	// Lock TTL — should be longer than max expected execution time
	lockTTLActivation = 50 * time.Second
	lockTTLDetection  = 2 * time.Minute
)

// GameSchedulerService handles cron-triggered game activation and match detection
type GameSchedulerService struct {
	gameDBPort        port.GameDatabasePort
	matchDetectionSvc *MatchDetectionService
	eventPublisher    port.GameEventPublisherPort
	redisClient       *redis.Client
}

func NewGameSchedulerService(
	gameDBPort port.GameDatabasePort,
	matchDetectionSvc *MatchDetectionService,
	eventPublisher port.GameEventPublisherPort,
	redisClient *redis.Client,
) *GameSchedulerService {
	return &GameSchedulerService{
		gameDBPort:        gameDBPort,
		matchDetectionSvc: matchDetectionSvc,
		eventPublisher:    eventPublisher,
		redisClient:       redisClient,
	}
}

// RunScheduledActivation is called every 1 minute by cron.
// It activates games whose scheduled start time has arrived.
func (s *GameSchedulerService) RunScheduledActivation() {
	ctx := context.Background()

	// Acquire distributed lock to prevent duplicate execution across instances
	acquired, err := s.acquireLock(ctx, lockKeyActivation, lockTTLActivation)
	if err != nil {
		log.Printf("[Scheduler] Failed to acquire activation lock: %v", err)
		return
	}
	if !acquired {
		log.Printf("[Scheduler] Activation job already running on another instance, skipping")
		return
	}
	defer s.releaseLock(ctx, lockKeyActivation)

	games, err := s.gameDBPort.GetGamesReadyToStart()
	if err != nil {
		log.Printf("[Scheduler] Failed to query games ready to start: %v", err)
		return
	}

	if len(games) == 0 {
		return
	}

	log.Printf("[Scheduler] Found %d games ready to activate", len(games))

	for _, game := range games {
		if err := game.ActivateForDetection(); err != nil {
			log.Printf("[Scheduler] Failed to activate game %d: %v", game.GameID, err)
			continue
		}

		if err := s.gameDBPort.Update(game); err != nil {
			log.Printf("[Scheduler] Failed to save activated game %d: %v", game.GameID, err)
			continue
		}

		// Publish activation event
		event := &port.GameEvent{
			EventType:   port.GameEventActivated,
			Timestamp:   time.Now(),
			ContestID:   game.ContestID,
			GameID:      game.GameID,
			Round:       game.GetRound(),
			MatchNumber: game.GetMatchNumber(),
		}
		if err := s.eventPublisher.PublishGameEvent(ctx, event); err != nil {
			log.Printf("[Scheduler] Failed to publish activation event for game %d: %v", game.GameID, err)
		}

		log.Printf("[Scheduler] Game %d activated (contest %d, round %d, match %d)",
			game.GameID, game.ContestID, game.GetRound(), game.GetMatchNumber())
	}
}

// RunMatchDetection is called every 3 minutes by cron.
// It runs match detection for all games currently in DETECTING state.
func (s *GameSchedulerService) RunMatchDetection() {
	ctx := context.Background()

	// Acquire distributed lock
	acquired, err := s.acquireLock(ctx, lockKeyDetection, lockTTLDetection)
	if err != nil {
		log.Printf("[Scheduler] Failed to acquire detection lock: %v", err)
		return
	}
	if !acquired {
		log.Printf("[Scheduler] Detection job already running on another instance, skipping")
		return
	}
	defer s.releaseLock(ctx, lockKeyDetection)

	games, err := s.gameDBPort.GetGamesInDetection()
	if err != nil {
		log.Printf("[Scheduler] Failed to query games in detection: %v", err)
		return
	}

	if len(games) == 0 {
		return
	}

	log.Printf("[Scheduler] Running match detection for %d games", len(games))

	for _, game := range games {
		// Skip games where API data is not yet available (Valorant API has ~30min delay)
		if game.ScheduledStartTime != nil {
			apiAvailableTime := game.ScheduledStartTime.Add(30 * time.Minute)
			if time.Now().Before(apiAvailableTime) {
				log.Printf("[Scheduler] Skipping game %d: API data not yet available (available after %s)",
					game.GameID, apiAvailableTime.Format(time.RFC3339))
				continue
			}
		}

		if err := s.matchDetectionSvc.DetectMatchForGame(game.GameID); err != nil {
			log.Printf("[Scheduler] Detection error for game %d: %v", game.GameID, err)
		}
	}
}

// acquireLock attempts to acquire a distributed lock using Redis SETNX
func (s *GameSchedulerService) acquireLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	result, err := s.redisClient.SetNX(ctx, key, fmt.Sprintf("locked:%d", time.Now().UnixMilli()), ttl).Result()
	if err != nil {
		return false, fmt.Errorf("redis SetNX failed: %w", err)
	}
	return result, nil
}

// releaseLock releases the distributed lock
func (s *GameSchedulerService) releaseLock(ctx context.Context, key string) {
	if err := s.redisClient.Del(ctx, key).Err(); err != nil {
		log.Printf("[Scheduler] Failed to release lock %s: %v", key, err)
	}
}
