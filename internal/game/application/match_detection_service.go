package application

import (
	"GAMERS-BE/internal/game/application/dto"
	"GAMERS-BE/internal/game/application/port"
	"GAMERS-BE/internal/game/domain"
	"GAMERS-BE/internal/global/exception"
	userQueryPort "GAMERS-BE/internal/user/application/port/port"
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// MatchDetectionService handles the detection of Valorant matches for tournament games
type MatchDetectionService struct {
	matchDetectionPort port.MatchDetectionPort
	gameDBPort         port.GameDatabasePort
	gameTeamDBPort     port.GameTeamDatabasePort
	teamDBPort         port.TeamDatabasePort
	matchResultDBPort  port.MatchResultDatabasePort
	eventPublisher     port.GameEventPublisherPort
	userQueryPort      userQueryPort.UserQueryPort
}

func NewMatchDetectionService(
	matchDetectionPort port.MatchDetectionPort,
	gameDBPort port.GameDatabasePort,
	gameTeamDBPort port.GameTeamDatabasePort,
	teamDBPort port.TeamDatabasePort,
	matchResultDBPort port.MatchResultDatabasePort,
	eventPublisher port.GameEventPublisherPort,
	userQueryPort userQueryPort.UserQueryPort,
) *MatchDetectionService {
	return &MatchDetectionService{
		matchDetectionPort: matchDetectionPort,
		gameDBPort:         gameDBPort,
		gameTeamDBPort:     gameTeamDBPort,
		teamDBPort:         teamDBPort,
		matchResultDBPort:  matchResultDBPort,
		eventPublisher:     eventPublisher,
		userQueryPort:      userQueryPort,
	}
}

// DetectMatchForGame runs match detection for a single game
func (s *MatchDetectionService) DetectMatchForGame(gameID int64) error {
	game, err := s.gameDBPort.GetByID(gameID)
	if err != nil {
		return err
	}

	if !game.IsDetecting() {
		return exception.ErrGameNotActive
	}

	// Check if detection window has expired
	if game.IsDetectionWindowExpired() {
		if err := game.MarkDetectionFailed(); err != nil {
			return err
		}
		if err := s.gameDBPort.Update(game); err != nil {
			return err
		}
		s.publishGameEvent(game, port.GameEventMatchFailed)
		log.Printf("[MatchDetection] Detection window expired for game %d", gameID)
		return nil
	}

	// Get the two teams participating in this game
	gameTeams, err := s.gameTeamDBPort.GetByGameID(gameID)
	if err != nil {
		return fmt.Errorf("failed to get game teams: %w", err)
	}
	if len(gameTeams) < 2 {
		return fmt.Errorf("game %d does not have 2 teams assigned", gameID)
	}

	teamA := gameTeams[0]
	teamB := gameTeams[1]

	// Get team members for both teams
	teamAMembers, err := s.teamDBPort.GetMembersByTeamID(teamA.TeamID)
	if err != nil {
		return fmt.Errorf("failed to get team A members: %w", err)
	}
	teamBMembers, err := s.teamDBPort.GetMembersByTeamID(teamB.TeamID)
	if err != nil {
		return fmt.Errorf("failed to get team B members: %w", err)
	}

	// Get Valorant account info for the first team's leader (used as reference for match lookup)
	leader, err := s.teamDBPort.GetLeaderByTeamID(teamA.TeamID)
	if err != nil {
		return fmt.Errorf("failed to get team leader: %w", err)
	}

	// Resolve Valorant name/tag for the leader
	// We need the Valorant account info stored in the user or oauth2 tables
	// For now, use the leader's linked Valorant info
	teamAInfo, err := s.getTeamValorantAccounts(teamA.TeamID, teamAMembers)
	if err != nil {
		return err
	}
	teamBInfo, err := s.getTeamValorantAccounts(teamB.TeamID, teamBMembers)
	if err != nil {
		return err
	}

	_ = leader // leader is used via teamAInfo

	if len(teamAInfo) == 0 {
		log.Printf("[MatchDetection] No Valorant accounts found for team %d in game %d", teamA.TeamID, gameID)
		return nil
	}

	// Use first member with a Valorant account to query match history
	refAccount := teamAInfo[0]
	matches, err := s.matchDetectionPort.GetRecentMatches("ap", refAccount.Name, refAccount.Tag)
	if err != nil {
		log.Printf("[MatchDetection] VAPI error for game %d: %v (will retry next cycle)", gameID, err)
		return nil // Non-fatal: retry on next polling cycle
	}

	// Filter matches within detection window
	windowStart := *game.ScheduledStartTime
	windowEnd := game.GetDetectionWindowEnd()

	var candidateMatches []port.ValorantMatch
	for _, m := range matches {
		if (m.GameStart.Equal(windowStart) || m.GameStart.After(windowStart)) &&
			(m.GameStart.Equal(windowEnd) || m.GameStart.Before(windowEnd)) {
			candidateMatches = append(candidateMatches, m)
		}
	}

	if len(candidateMatches) == 0 {
		log.Printf("[MatchDetection] No matches in window for game %d", gameID)
		return nil
	}

	// Check each candidate match (latest first) for full team participation
	// If multiple matches qualify, pick the latest one
	var bestMatch *port.ValorantMatchDetail
	for i := len(candidateMatches) - 1; i >= 0; i-- {
		detail, err := s.matchDetectionPort.GetMatchDetail(candidateMatches[i].MatchID)
		if err != nil {
			log.Printf("[MatchDetection] Failed to get match detail %s: %v", candidateMatches[i].MatchID, err)
			continue
		}

		if s.ValidateMatchParticipants(detail, teamAInfo, teamBInfo) {
			bestMatch = detail
			break
		}
	}

	if bestMatch == nil {
		log.Printf("[MatchDetection] No qualifying match found for game %d", gameID)
		return nil
	}

	// Process the detected match
	return s.ProcessDetectedMatch(game, bestMatch, teamA, teamB, teamAInfo, teamBInfo)
}

// ValorantAccountInfo holds a player's Valorant account details for matching
type ValorantAccountInfo struct {
	UserID int64
	TeamID int64
	Name   string
	Tag    string
}

// getTeamValorantAccounts resolves Valorant name/tag for all members of a team
// by looking up each member's linked Valorant account from the user table.
func (s *MatchDetectionService) getTeamValorantAccounts(teamID int64, members []*domain.TeamMember) ([]ValorantAccountInfo, error) {
	var accounts []ValorantAccountInfo
	for _, m := range members {
		user, err := s.userQueryPort.FindById(m.UserID)
		if err != nil {
			log.Printf("[MatchDetection] Failed to find user %d: %v", m.UserID, err)
			continue
		}
		if !user.HasValorantLinked() {
			log.Printf("[MatchDetection] User %d (team %d) has no Valorant account linked, skipping", m.UserID, teamID)
			continue
		}
		accounts = append(accounts, ValorantAccountInfo{
			UserID: m.UserID,
			TeamID: teamID,
			Name:   *user.RiotName,
			Tag:    *user.RiotTag,
		})
	}
	return accounts, nil
}

// ValidateMatchParticipants checks that all members of both teams are present in the match
func (s *MatchDetectionService) ValidateMatchParticipants(
	match *port.ValorantMatchDetail,
	teamAAccounts, teamBAccounts []ValorantAccountInfo,
) bool {
	if match == nil {
		return false
	}

	// Build a set of all match participants by normalized "name#tag"
	participantSet := make(map[string]bool, len(match.Players))
	for _, p := range match.Players {
		key := normalizeNameTag(p.Name, p.Tag)
		participantSet[key] = true
	}

	// Verify every member of team A is in the match
	for _, account := range teamAAccounts {
		key := normalizeNameTag(account.Name, account.Tag)
		if !participantSet[key] {
			return false
		}
	}

	// Verify every member of team B is in the match
	for _, account := range teamBAccounts {
		key := normalizeNameTag(account.Name, account.Tag)
		if !participantSet[key] {
			return false
		}
	}

	return true
}

// normalizeNameTag creates a lowercase "name#tag" key for comparison
func normalizeNameTag(name, tag string) string {
	return strings.ToLower(name) + "#" + strings.ToLower(tag)
}

// ProcessDetectedMatch records the match result, updates game state, and advances bracket
func (s *MatchDetectionService) ProcessDetectedMatch(
	game *domain.Game,
	match *port.ValorantMatchDetail,
	teamA, teamB *domain.GameTeam,
	teamAAccounts, teamBAccounts []ValorantAccountInfo,
) error {
	// Determine which Valorant team side each tournament team is on
	winnerTeamID, loserTeamID, winnerScore, loserScore := s.resolveWinner(
		match, teamA, teamB, teamAAccounts, teamBAccounts,
	)

	// Mark the game as detected
	if err := game.MarkDetected(match.MatchID); err != nil {
		return err
	}

	// Save match result
	matchResult := domain.NewMatchResult(
		game.GameID,
		match.MatchID,
		match.MapName,
		match.RoundsPlayed,
		winnerTeamID, loserTeamID,
		winnerScore, loserScore,
		match.GameStart,
		match.GameLength,
	)

	savedResult, err := s.matchResultDBPort.Save(matchResult)
	if err != nil {
		return fmt.Errorf("failed to save match result: %w", err)
	}

	// Save player stats
	playerStats := s.buildPlayerStats(savedResult.MatchResultID, match, teamAAccounts, teamBAccounts)
	if len(playerStats) > 0 {
		if err := s.matchResultDBPort.SavePlayerStats(playerStats); err != nil {
			log.Printf("[MatchDetection] Failed to save player stats for game %d: %v", game.GameID, err)
		}
	}

	// Update GameTeam grades: winner=1, loser=2
	s.updateGameTeamGrades(teamA, teamB, winnerTeamID)

	// Finish the game
	if err := game.FinishGame(); err != nil {
		return err
	}
	if err := s.gameDBPort.Update(game); err != nil {
		return err
	}

	// Advance winner to next round if applicable
	if game.NextGameID != nil {
		s.advanceWinnerToNextGame(*game.NextGameID, winnerTeamID)
	}

	// Publish events
	s.publishMatchDetectedEvent(game, match, winnerTeamID, loserTeamID, winnerScore, loserScore)
	s.publishGameEvent(game, port.GameEventFinished)

	log.Printf("[MatchDetection] Game %d finished. Winner: team %d, Score: %d-%d",
		game.GameID, winnerTeamID, winnerScore, loserScore)

	return nil
}

// resolveWinner determines which tournament team won based on match player team assignments
func (s *MatchDetectionService) resolveWinner(
	match *port.ValorantMatchDetail,
	teamA, teamB *domain.GameTeam,
	teamAAccounts, teamBAccounts []ValorantAccountInfo,
) (winnerTeamID, loserTeamID int64, winnerScore, loserScore int) {
	// Map tournament team -> Valorant side (Red/Blue)
	teamASide := s.resolveTeamSide(match, teamAAccounts)

	// Find which Valorant side won
	var winningSide string
	for _, t := range match.Teams {
		if t.HasWon {
			winningSide = t.TeamID
			break
		}
	}

	// Get scores by side
	sideScores := make(map[string]int)
	for _, t := range match.Teams {
		sideScores[t.TeamID] = t.RoundsWon
	}

	if teamASide == winningSide {
		return teamA.TeamID, teamB.TeamID, sideScores[teamASide], sideScores[s.otherSide(teamASide)]
	}
	otherSide := s.otherSide(teamASide)
	return teamB.TeamID, teamA.TeamID, sideScores[otherSide], sideScores[teamASide]
}

// resolveTeamSide determines which Valorant side (Red/Blue) a tournament team is on
func (s *MatchDetectionService) resolveTeamSide(match *port.ValorantMatchDetail, teamAccounts []ValorantAccountInfo) string {
	if len(teamAccounts) == 0 {
		return ""
	}
	refKey := normalizeNameTag(teamAccounts[0].Name, teamAccounts[0].Tag)
	for _, p := range match.Players {
		if normalizeNameTag(p.Name, p.Tag) == refKey {
			return p.TeamID
		}
	}
	return ""
}

func (s *MatchDetectionService) otherSide(side string) string {
	if side == "Red" {
		return "Blue"
	}
	return "Red"
}

// buildPlayerStats constructs MatchPlayerStat entries from match data
func (s *MatchDetectionService) buildPlayerStats(
	matchResultID int64,
	match *port.ValorantMatchDetail,
	teamAAccounts, teamBAccounts []ValorantAccountInfo,
) []*domain.MatchPlayerStat {
	// Build lookup: normalized name#tag -> ValorantAccountInfo
	accountMap := make(map[string]ValorantAccountInfo)
	for _, a := range teamAAccounts {
		accountMap[normalizeNameTag(a.Name, a.Tag)] = a
	}
	for _, a := range teamBAccounts {
		accountMap[normalizeNameTag(a.Name, a.Tag)] = a
	}

	var stats []*domain.MatchPlayerStat
	for _, p := range match.Players {
		key := normalizeNameTag(p.Name, p.Tag)
		account, found := accountMap[key]
		if !found {
			continue
		}
		stat := domain.NewMatchPlayerStat(
			matchResultID, account.UserID, account.TeamID,
			p.Agent,
			p.Kills, p.Deaths, p.Assists, p.Score,
			p.Headshots, p.Bodyshots, p.Legshots,
		)
		stats = append(stats, stat)
	}
	return stats
}

func (s *MatchDetectionService) updateGameTeamGrades(teamA, teamB *domain.GameTeam, winnerTeamID int64) {
	if teamA.TeamID == winnerTeamID {
		teamA.SetGrade(1)
		teamB.SetGrade(2)
	} else {
		teamA.SetGrade(2)
		teamB.SetGrade(1)
	}
	// Grades are persisted via GameTeam updates (caller should handle persistence)
}

func (s *MatchDetectionService) advanceWinnerToNextGame(nextGameID, winnerTeamID int64) {
	nextGameTeam := domain.NewGameTeam(nextGameID, winnerTeamID)
	if _, err := s.gameTeamDBPort.Save(nextGameTeam); err != nil {
		log.Printf("[MatchDetection] Failed to advance team %d to next game %d: %v",
			winnerTeamID, nextGameID, err)
	}
}

func (s *MatchDetectionService) publishGameEvent(game *domain.Game, eventType port.GameEventType) {
	event := &port.GameEvent{
		EventType: eventType,
		Timestamp: time.Now(),
		ContestID: game.ContestID,
		GameID:    game.GameID,
		Round:     game.GetRound(),
		MatchNumber: game.GetMatchNumber(),
	}
	if err := s.eventPublisher.PublishGameEvent(context.Background(), event); err != nil {
		log.Printf("[MatchDetection] Failed to publish %s event for game %d: %v",
			eventType, game.GameID, err)
	}
}

func (s *MatchDetectionService) publishMatchDetectedEvent(
	game *domain.Game,
	match *port.ValorantMatchDetail,
	winnerTeamID, loserTeamID int64,
	winnerScore, loserScore int,
) {
	event := &port.MatchDetectedEvent{
		GameEvent: port.GameEvent{
			EventType:   port.GameEventMatchDetected,
			Timestamp:   time.Now(),
			ContestID:   game.ContestID,
			GameID:      game.GameID,
			Round:       game.GetRound(),
			MatchNumber: game.GetMatchNumber(),
		},
		ValorantMatchID: match.MatchID,
		WinnerTeamID:    winnerTeamID,
		LoserTeamID:     loserTeamID,
		Score:           fmt.Sprintf("%d-%d", winnerScore, loserScore),
		MapName:         match.MapName,
	}
	if err := s.eventPublisher.PublishMatchDetectedEvent(context.Background(), event); err != nil {
		log.Printf("[MatchDetection] Failed to publish match detected event for game %d: %v",
			game.GameID, err)
	}
}

// SubmitManualResult allows staff to manually input a game result
func (s *MatchDetectionService) SubmitManualResult(gameID int64, req *dto.ManualResultRequest) (*domain.MatchResult, error) {
	game, err := s.gameDBPort.GetByID(gameID)
	if err != nil {
		return nil, err
	}

	// Manual result allowed when DETECTING or FAILED
	if game.DetectionStatus != domain.DetectionStatusDetecting &&
		game.DetectionStatus != domain.DetectionStatusFailed &&
		game.DetectionStatus != domain.DetectionStatusNone {
		return nil, exception.ErrDetectionNotFailed
	}

	// Validate winner team is part of this game
	gameTeams, err := s.gameTeamDBPort.GetByGameID(gameID)
	if err != nil {
		return nil, err
	}

	var winnerGT, loserGT *domain.GameTeam
	for _, gt := range gameTeams {
		if gt.TeamID == req.WinnerTeamID {
			winnerGT = gt
		} else {
			loserGT = gt
		}
	}
	if winnerGT == nil {
		return nil, exception.ErrWinnerTeamNotInGame
	}
	if loserGT == nil {
		return nil, exception.ErrWinnerTeamNotInGame
	}

	// Mark as manual
	if err := game.MarkManualResult(); err != nil {
		// If state doesn't allow transition, force it for manual override
		game.DetectionStatus = domain.DetectionStatusManual
	}

	// Create match result (no Valorant match ID for manual)
	matchResult := domain.NewMatchResult(
		gameID,
		"manual",
		"",
		req.WinnerScore+req.LoserScore,
		req.WinnerTeamID, loserGT.TeamID,
		req.WinnerScore, req.LoserScore,
		time.Now(),
		0,
	)

	savedResult, err := s.matchResultDBPort.Save(matchResult)
	if err != nil {
		return nil, fmt.Errorf("failed to save manual result: %w", err)
	}

	// Update grades
	winnerGT.SetGrade(1)
	loserGT.SetGrade(2)

	// Finish the game
	if game.GameStatus == domain.GameStatusActive {
		if err := game.FinishGame(); err != nil {
			return nil, err
		}
	} else if game.GameStatus == domain.GameStatusPending {
		// For pending games that were never activated, transition through states
		now := time.Now()
		game.GameStatus = domain.GameStatusFinished
		game.StartedAt = &now
		game.EndedAt = &now
		game.ModifiedAt = now
	}

	if err := s.gameDBPort.Update(game); err != nil {
		return nil, err
	}

	// Advance winner to next round
	if game.NextGameID != nil {
		s.advanceWinnerToNextGame(*game.NextGameID, req.WinnerTeamID)
	}

	s.publishGameEvent(game, port.GameEventManualResult)
	s.publishGameEvent(game, port.GameEventFinished)

	return savedResult, nil
}

// GetMatchResult returns the match result for a game
func (s *MatchDetectionService) GetMatchResult(gameID int64) (*dto.MatchResultResponse, error) {
	game, err := s.gameDBPort.GetByID(gameID)
	if err != nil {
		return nil, err
	}

	result, err := s.matchResultDBPort.GetByGameID(gameID)
	if err != nil {
		return nil, err
	}

	resp := dto.ToMatchResultResponse(result, game.DetectionStatus)
	return resp, nil
}

// GetMatchResultWithStats returns the match result with player stats
func (s *MatchDetectionService) GetMatchResultWithStats(gameID int64) (*dto.MatchResultResponse, error) {
	resp, err := s.GetMatchResult(gameID)
	if err != nil {
		return nil, err
	}

	stats, err := s.matchResultDBPort.GetPlayerStatsByMatchResult(resp.MatchResultID)
	if err != nil {
		return nil, err
	}

	playerStats := make([]*dto.PlayerStatResponse, len(stats))
	for i, st := range stats {
		playerStats[i] = dto.ToPlayerStatResponse(st)
	}
	resp.PlayerStats = playerStats

	return resp, nil
}
