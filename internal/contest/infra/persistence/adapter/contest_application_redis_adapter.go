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

// RequestParticipate - 신청 요청
func (c *ContestApplicationRedisAdapter) RequestParticipate(ctx context.Context, contestId, userId int64, ttl time.Duration) error {
	// 중복 신청 확인
	hasApplied, err := c.HasApplied(ctx, contestId, userId)
	if err != nil {
		return err
	}
	if hasApplied {
		return exception.ErrAlreadyApplied
	}

	pipe := c.client.Pipeline()

	// 1. 신청 정보 저장 (Hash)
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

	// 2. Pending 목록에 추가 (Sorted Set - 신청 시간으로 정렬)
	pendingKey := utils.GetPendingKey(contestId)
	pipe.ZAdd(ctx, pendingKey, redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: userId,
	})
	pipe.Expire(ctx, pendingKey, ttl)

	// 3. User의 신청 목록에 추가
	userAppKey := utils.GetUserApplicationsKey(userId)
	pipe.SAdd(ctx, userAppKey, contestId)
	pipe.Expire(ctx, userAppKey, 30*24*time.Hour) // 30일 유지

	_, err = pipe.Exec(ctx)
	return err
}

// AcceptRequest - 신청 승인
func (c *ContestApplicationRedisAdapter) AcceptRequest(ctx context.Context, contestId, userId, processedBy int64) error {
	// 신청 정보 조회
	app, err := c.GetApplication(ctx, contestId, userId)
	if err != nil {
		return err
	}

	if app.Status != port.ApplicationStatusPending {
		return exception.ErrApplicationNotPending
	}

	pipe := c.client.Pipeline()

	// 1. 신청 상태 업데이트
	appKey := utils.GetApplicationKey(contestId, userId)
	now := time.Now()
	app.Status = port.ApplicationStatusAccepted
	app.ProcessedAt = &now
	app.ProcessedBy = &processedBy

	appData, err := json.Marshal(app)
	if err != nil {
		return err
	}

	// 기존 TTL 유지
	ttl := c.client.TTL(ctx, appKey).Val()
	pipe.Set(ctx, appKey, appData, ttl)

	// 2. Pending에서 제거
	pendingKey := utils.GetPendingKey(contestId)
	pipe.ZRem(ctx, pendingKey, userId)

	// 3. Accepted Set에 추가
	acceptedKey := utils.GetAcceptedKey(contestId)
	pipe.SAdd(ctx, acceptedKey, userId)
	pipe.Expire(ctx, acceptedKey, ttl)

	_, err = pipe.Exec(ctx)
	return err
}

// RejectRequest - 신청 거절
func (c *ContestApplicationRedisAdapter) RejectRequest(ctx context.Context, contestId, userId, processedBy int64) error {
	// 신청 정보 조회
	app, err := c.GetApplication(ctx, contestId, userId)
	if err != nil {
		return err
	}

	if app.Status != port.ApplicationStatusPending {
		return exception.ErrApplicationNotPending
	}

	pipe := c.client.Pipeline()

	// 1. 신청 상태 업데이트
	appKey := utils.GetApplicationKey(contestId, userId)
	now := time.Now()
	app.Status = port.ApplicationStatusRejected
	app.ProcessedAt = &now
	app.ProcessedBy = &processedBy

	appData, err := json.Marshal(app)
	if err != nil {
		return err
	}

	// 기존 TTL 유지
	ttl := c.client.TTL(ctx, appKey).Val()
	pipe.Set(ctx, appKey, appData, ttl)

	// 2. Pending에서 제거
	pendingKey := utils.GetPendingKey(contestId)
	pipe.ZRem(ctx, pendingKey, userId)

	// 3. Rejected Set에 추가
	rejectedKey := utils.GetRejectedKey(contestId)
	pipe.SAdd(ctx, rejectedKey, userId)
	pipe.Expire(ctx, rejectedKey, ttl)

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
func (c *ContestApplicationRedisAdapter) HasApplied(ctx context.Context, contestId, userId int64) (bool, error) {
	userAppKey := utils.GetUserApplicationsKey(userId)
	exists, err := c.client.SIsMember(ctx, userAppKey, contestId).Result()
	if err != nil {
		return false, err
	}
	return exists, nil
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
