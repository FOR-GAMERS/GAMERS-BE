package adapter

import (
	"GAMERS-BE/internal/contest/application/port"
	"GAMERS-BE/internal/global/exception"
	"GAMERS-BE/internal/global/utils"
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type ContestApplicationRedisAdapter struct {
	client *redis.Client
}

func NewContestApplicationRedisAdapter(client *redis.Client) *ContestApplicationRedisAdapter {
	return &ContestApplicationRedisAdapter{
		client: client,
	}
}

func (c *ContestApplicationRedisAdapter) RequestParticipate(ctx context.Context, contestId, userId int64, ttl time.Duration) error {
	hasApplied, err := c.HasApplied(ctx, contestId, userId)
	if err != nil {
		return err
	}
	if hasApplied {
		return exception.ErrAlreadyApplied
	}

	pipe := c.client.Pipeline()

	appKey := utils.GetApplicationKey(contestId, userId)
	application := port.ContestApplication{
		UserID:      userId,
		ContestID:   contestId,
		Status:      port.ApplicationStatusPending,
		RequestedAt: time.Now(),
	}

	appData, err := json.Marshal(application)
	if err != nil {
		return err
	}

	pipe.Set(ctx, appKey, appData, ttl)

	pendingKey := utils.GetPendingKey(contestId)
	pipe.ZAdd(ctx, pendingKey, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: userId,
	})
	pipe.Expire(ctx, pendingKey, ttl)

	userAppKey := utils.GetUserApplicationsKey(userId)
	pipe.SAdd(ctx, userAppKey, contestId)
	pipe.Expire(ctx, userAppKey, 30*24*time.Hour)

	_, err = pipe.Exec(ctx)
	return err
}

func (c *ContestApplicationRedisAdapter) AcceptRequest(ctx context.Context, contestId, userId, processedBy int64) error {
	app, err := c.GetApplication(ctx, contestId, userId)
	if err != nil {
		return err
	}

	if app.Status != port.ApplicationStatusPending {
		return exception.ErrApplicationNotPending
	}

	pipe := c.client.Pipeline()

	appKey := utils.GetApplicationKey(contestId, userId)
	now := time.Now()
	app.Status = port.ApplicationStatusAccepted
	app.ProcessedAt = &now
	app.ProcessedBy = &processedBy

	appData, err := json.Marshal(app)
	if err != nil {
		return err
	}

	ttl := c.client.TTL(ctx, appKey).Val()
	pipe.Set(ctx, appKey, appData, ttl)

	pendingKey := utils.GetPendingKey(contestId)
	pipe.ZRem(ctx, pendingKey, userId)

	acceptedKey := utils.GetAcceptedKey(contestId)
	pipe.SAdd(ctx, acceptedKey, userId)
	pipe.Expire(ctx, acceptedKey, ttl)

	_, err = pipe.Exec(ctx)
	return err
}

func (c *ContestApplicationRedisAdapter) CancelApplication(ctx context.Context, contestId, userId int64) error {
	app, err := c.GetApplication(ctx, contestId, userId)
	if err != nil {
		return err
	}

	if app.Status != port.ApplicationStatusPending {
		return exception.ErrApplicationNotPending
	}

	pipe := c.client.Pipeline()

	appKey := utils.GetApplicationKey(contestId, userId)
	pipe.Del(ctx, appKey)

	pendingKey := utils.GetPendingKey(contestId)
	pipe.ZRem(ctx, pendingKey, userId)

	userAppKey := utils.GetUserApplicationsKey(userId)
	pipe.SRem(ctx, userAppKey, contestId)

	_, err = pipe.Exec(ctx)
	return err
}

func (c *ContestApplicationRedisAdapter) RejectRequest(ctx context.Context, contestId, userId, processedBy int64) error {
	app, err := c.GetApplication(ctx, contestId, userId)
	if err != nil {
		return err
	}

	if app.Status != port.ApplicationStatusPending {
		return exception.ErrApplicationNotPending
	}

	pipe := c.client.Pipeline()

	appKey := utils.GetApplicationKey(contestId, userId)
	now := time.Now()
	app.Status = port.ApplicationStatusRejected
	app.ProcessedAt = &now
	app.ProcessedBy = &processedBy

	appData, err := json.Marshal(app)
	if err != nil {
		return err
	}

	ttl := c.client.TTL(ctx, appKey).Val()
	pipe.Set(ctx, appKey, appData, ttl)

	pendingKey := utils.GetPendingKey(contestId)
	pipe.ZRem(ctx, pendingKey, userId)

	rejectedKey := utils.GetRejectedKey(contestId)
	pipe.SAdd(ctx, rejectedKey, userId)
	pipe.Expire(ctx, rejectedKey, ttl)

	userAppKey := utils.GetUserApplicationsKey(userId)
	pipe.SRem(ctx, userAppKey, contestId)

	_, err = pipe.Exec(ctx)
	return err
}

// GetApplication - 신청 정보 조회
func (c *ContestApplicationRedisAdapter) GetApplication(ctx context.Context, contestId, userId int64) (*port.ContestApplication, error) {
	appKey := utils.GetApplicationKey(contestId, userId)
	data, err := c.client.Get(ctx, appKey).Result()
	if errors.Is(err, redis.Nil) {
		return nil, exception.ErrApplicationNotFound
	}
	if err != nil {
		return nil, err
	}

	var app port.ContestApplication
	if err := json.Unmarshal([]byte(data), &app); err != nil {
		return nil, err
	}

	return &app, nil
}

// GetPendingApplications - Pending 신청 목록 조회
func (c *ContestApplicationRedisAdapter) GetPendingApplications(ctx context.Context, contestId int64) ([]*port.ContestApplication, error) {
	pendingKey := utils.GetPendingKey(contestId)

	// Sorted Set에서 모든 userId 조회 (오래된 순)
	userIDs, err := c.client.ZRange(ctx, pendingKey, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	applications := make([]*port.ContestApplication, 0, len(userIDs))
	for _, userIDStr := range userIDs {
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			continue
		}

		app, err := c.GetApplication(ctx, contestId, userID)
		if err != nil {
			continue // 신청 정보가 없으면 스킵
		}

		applications = append(applications, app)
	}

	return applications, nil
}

// GetAcceptedApplications - Accepted 신청 목록 조회
func (c *ContestApplicationRedisAdapter) GetAcceptedApplications(ctx context.Context, contestId int64) ([]int64, error) {
	acceptedKey := utils.GetAcceptedKey(contestId)
	members, err := c.client.SMembers(ctx, acceptedKey).Result()
	if err != nil {
		return nil, err
	}

	userIDs := make([]int64, 0, len(members))
	for _, member := range members {
		userID, err := strconv.ParseInt(member, 10, 64)
		if err != nil {
			continue
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

// GetRejectedApplications - Rejected 신청 목록 조회
func (c *ContestApplicationRedisAdapter) GetRejectedApplications(ctx context.Context, contestId int64) ([]int64, error) {
	rejectedKey := utils.GetRejectedKey(contestId)
	members, err := c.client.SMembers(ctx, rejectedKey).Result()
	if err != nil {
		return nil, err
	}

	userIDs := make([]int64, 0, len(members))
	for _, member := range members {
		userID, err := strconv.ParseInt(member, 10, 64)
		if err != nil {
			continue
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

// HasApplied - 중복 신청 확인
// Pending/Accepted 상태면 재신청 불가, Rejected 상태면 재신청 가능
func (c *ContestApplicationRedisAdapter) HasApplied(ctx context.Context, contestId, userId int64) (bool, error) {
	app, err := c.GetApplication(ctx, contestId, userId)
	if err != nil {
		if err == exception.ErrApplicationNotFound {
			return false, nil
		}
		return false, err
	}
	// Rejected 상태는 재신청 가능
	if app.Status == port.ApplicationStatusRejected {
		return false, nil
	}
	return true, nil
}

// ExtendTTL - TTL 연장
func (c *ContestApplicationRedisAdapter) ExtendTTL(ctx context.Context, contestId int64, newTTL time.Duration) error {
	pattern := utils.GetContestPatternKey(contestId)

	var cursor uint64
	for {
		keys, nextCursor, err := c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}

		pipe := c.client.Pipeline()
		for _, key := range keys {
			pipe.Expire(ctx, key, newTTL)
		}
		_, err = pipe.Exec(ctx)
		if err != nil {
			return err
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}

// ClearApplications - Contest의 모든 신청 정보 삭제
func (c *ContestApplicationRedisAdapter) ClearApplications(ctx context.Context, contestId int64) error {
	pattern := utils.GetContestPatternKey(contestId)

	var cursor uint64
	for {
		keys, nextCursor, err := c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			pipe := c.client.Pipeline()
			for _, key := range keys {
				pipe.Del(ctx, key)
			}
			_, err = pipe.Exec(ctx)
			if err != nil {
				return err
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}
