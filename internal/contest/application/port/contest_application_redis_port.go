package port

import (
	"context"
	"time"
)

type ApplicationStatus string

const (
	ApplicationStatusPending  ApplicationStatus = "PENDING"
	ApplicationStatusAccepted ApplicationStatus = "ACCEPTED"
	ApplicationStatusRejected ApplicationStatus = "REJECTED"
)

type ContestApplication struct {
	UserID      int64             `json:"user_id"`
	ContestID   int64             `json:"contest_id"`
	Status      ApplicationStatus `json:"status"`
	RequestedAt time.Time         `json:"requested_at"`
	ProcessedAt *time.Time        `json:"processed_at,omitempty"`
	ProcessedBy *int64            `json:"processed_by,omitempty"`
}

type ContestApplicationRedisPort interface {
	// 신청 관리
	RequestParticipate(ctx context.Context, contestId, userId int64, ttl time.Duration) error
	AcceptRequest(ctx context.Context, contestId, userId, processedBy int64) error
	RejectRequest(ctx context.Context, contestId, userId, processedBy int64) error
	CancelApplication(ctx context.Context, contestId, userId int64) error

	// 조회
	GetApplication(ctx context.Context, contestId, userId int64) (*ContestApplication, error)
	GetPendingApplications(ctx context.Context, contestId int64) ([]*ContestApplication, error)
	GetAcceptedApplications(ctx context.Context, contestId int64) ([]int64, error)
	GetRejectedApplications(ctx context.Context, contestId int64) ([]int64, error)

	// 중복 확인
	HasApplied(ctx context.Context, contestId, userId int64) (bool, error)

	// TTL 관리
	ExtendTTL(ctx context.Context, contestId int64, newTTL time.Duration) error

	// 정리
	ClearApplications(ctx context.Context, contestId int64) error
}
