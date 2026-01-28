package port

import (
	"time"
)

// TeamPersistenceEventType represents the type of team persistence event
type TeamPersistenceEventType string

const (
	TeamPersistenceCreated       TeamPersistenceEventType = "team.persistence.created"
	TeamPersistenceMemberAdded   TeamPersistenceEventType = "team.persistence.member_added"
	TeamPersistenceMemberRemoved TeamPersistenceEventType = "team.persistence.member_removed"
	TeamPersistenceFinalized     TeamPersistenceEventType = "team.persistence.finalized"
	TeamPersistenceDeleted       TeamPersistenceEventType = "team.persistence.deleted"
)

// TeamMemberPersistence represents member data for persistence
type TeamMemberPersistence struct {
	UserID     int64          `json:"user_id"`
	MemberType TeamMemberType `json:"member_type"`
	JoinedAt   time.Time      `json:"joined_at"`
}

// TeamPersistenceEvent represents an event for async DB persistence
type TeamPersistenceEvent struct {
	EventID    string                   `json:"event_id"`
	EventType  TeamPersistenceEventType `json:"event_type"`
	Timestamp  time.Time                `json:"timestamp"`
	RetryCount int                      `json:"retry_count"`
	ContestID  int64                    `json:"contest_id"`
	TeamID     int64                    `json:"team_id"`
	TeamName   *string                  `json:"team_name,omitempty"`
	Members    []*TeamMemberPersistence `json:"members,omitempty"`
	// For single member operations
	MemberUserID   *int64          `json:"member_user_id,omitempty"`
	MemberType     *TeamMemberType `json:"member_type,omitempty"`
	MemberJoinedAt *time.Time      `json:"member_joined_at,omitempty"`
}

const (
	// MaxRetryCount is the maximum number of retries before sending to DLQ
	MaxRetryCount = 3
)
