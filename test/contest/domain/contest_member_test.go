package domain_test

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/contest/domain"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/exception"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemberType_IsValid(t *testing.T) {
	tests := []struct {
		name       string
		memberType domain.MemberType
		expected   bool
	}{
		{"STAFF is valid", domain.MemberTypeStaff, true},
		{"NORMAL is valid", domain.MemberTypeNormal, true},
		{"empty string is invalid", domain.MemberType(""), false},
		{"unknown type is invalid", domain.MemberType("UNKNOWN"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.memberType.IsValid())
		})
	}
}

func TestNewContestMemberAsLeader(t *testing.T) {
	t.Run("creates leader with correct attributes", func(t *testing.T) {
		userID := int64(1)
		contestID := int64(100)

		member := domain.NewContestMemberAsLeader(userID, contestID)

		assert.NotNil(t, member)
		assert.Equal(t, userID, member.UserID)
		assert.Equal(t, contestID, member.ContestID)
		assert.Equal(t, domain.MemberTypeStaff, member.MemberType)
		assert.Equal(t, domain.LeaderTypeLeader, member.LeaderType)
		assert.Equal(t, 0, member.Point)
	})
}

func TestNewContestMember(t *testing.T) {
	t.Run("creates regular member with specified attributes", func(t *testing.T) {
		userID := int64(2)
		contestID := int64(100)

		member := domain.NewContestMember(userID, contestID, domain.MemberTypeNormal, domain.LeaderTypeMember)

		assert.NotNil(t, member)
		assert.Equal(t, userID, member.UserID)
		assert.Equal(t, contestID, member.ContestID)
		assert.Equal(t, domain.MemberTypeNormal, member.MemberType)
		assert.Equal(t, domain.LeaderTypeMember, member.LeaderType)
		assert.Equal(t, 0, member.Point)
	})

	t.Run("creates staff member with specified attributes", func(t *testing.T) {
		userID := int64(3)
		contestID := int64(100)

		member := domain.NewContestMember(userID, contestID, domain.MemberTypeStaff, domain.LeaderTypeMember)

		assert.NotNil(t, member)
		assert.Equal(t, domain.MemberTypeStaff, member.MemberType)
		assert.Equal(t, domain.LeaderTypeMember, member.LeaderType)
	})
}

func TestContestMember_TableName(t *testing.T) {
	member := &domain.ContestMember{}
	assert.Equal(t, "contests_members", member.TableName())
}

func TestContestMember_IsLeader(t *testing.T) {
	tests := []struct {
		name       string
		leaderType domain.LeaderType
		expected   bool
	}{
		{"LEADER returns true", domain.LeaderTypeLeader, true},
		{"MEMBER returns false", domain.LeaderTypeMember, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			member := &domain.ContestMember{LeaderType: tt.leaderType}
			assert.Equal(t, tt.expected, member.IsLeader())
		})
	}
}

func TestContestMember_IsStaff(t *testing.T) {
	tests := []struct {
		name       string
		memberType domain.MemberType
		expected   bool
	}{
		{"STAFF returns true", domain.MemberTypeStaff, true},
		{"NORMAL returns false", domain.MemberTypeNormal, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			member := &domain.ContestMember{MemberType: tt.memberType}
			assert.Equal(t, tt.expected, member.IsStaff())
		})
	}
}

func TestContestMember_IsValidMemberType(t *testing.T) {
	tests := []struct {
		name       string
		memberType domain.MemberType
		expected   bool
	}{
		{"STAFF is valid", domain.MemberTypeStaff, true},
		{"NORMAL is valid", domain.MemberTypeNormal, true},
		{"empty string is invalid", domain.MemberType(""), false},
		{"unknown type is invalid", domain.MemberType("UNKNOWN"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			member := &domain.ContestMember{MemberType: tt.memberType}
			assert.Equal(t, tt.expected, member.IsValidMemberType())
		})
	}
}

func TestContestMember_IsValidLeaderType(t *testing.T) {
	tests := []struct {
		name       string
		leaderType domain.LeaderType
		expected   bool
	}{
		{"LEADER is valid", domain.LeaderTypeLeader, true},
		{"MEMBER is valid", domain.LeaderTypeMember, true},
		{"empty string is invalid", domain.LeaderType(""), false},
		{"unknown type is invalid", domain.LeaderType("UNKNOWN"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			member := &domain.ContestMember{LeaderType: tt.leaderType}
			assert.Equal(t, tt.expected, member.IsValidLeaderType())
		})
	}
}

func TestContestMember_Validate(t *testing.T) {
	t.Run("valid member passes validation", func(t *testing.T) {
		member := &domain.ContestMember{
			UserID:     1,
			ContestID:  100,
			MemberType: domain.MemberTypeNormal,
			LeaderType: domain.LeaderTypeMember,
			Point:      0,
		}
		err := member.Validate()
		assert.NoError(t, err)
	})

	t.Run("valid leader passes validation", func(t *testing.T) {
		member := &domain.ContestMember{
			UserID:     1,
			ContestID:  100,
			MemberType: domain.MemberTypeStaff,
			LeaderType: domain.LeaderTypeLeader,
			Point:      50,
		}
		err := member.Validate()
		assert.NoError(t, err)
	})

	t.Run("fails with zero UserID", func(t *testing.T) {
		member := &domain.ContestMember{
			UserID:     0,
			ContestID:  100,
			MemberType: domain.MemberTypeNormal,
			LeaderType: domain.LeaderTypeMember,
			Point:      0,
		}
		err := member.Validate()
		assert.Equal(t, exception.ErrInvalidUserID, err)
	})

	t.Run("fails with zero ContestID", func(t *testing.T) {
		member := &domain.ContestMember{
			UserID:     1,
			ContestID:  0,
			MemberType: domain.MemberTypeNormal,
			LeaderType: domain.LeaderTypeMember,
			Point:      0,
		}
		err := member.Validate()
		assert.Equal(t, exception.ErrInvalidContestID, err)
	})

	t.Run("fails with invalid MemberType", func(t *testing.T) {
		member := &domain.ContestMember{
			UserID:     1,
			ContestID:  100,
			MemberType: domain.MemberType("INVALID"),
			LeaderType: domain.LeaderTypeMember,
			Point:      0,
		}
		err := member.Validate()
		assert.Equal(t, exception.ErrInvalidMemberType, err)
	})

	t.Run("fails with invalid LeaderType", func(t *testing.T) {
		member := &domain.ContestMember{
			UserID:     1,
			ContestID:  100,
			MemberType: domain.MemberTypeNormal,
			LeaderType: domain.LeaderType("INVALID"),
			Point:      0,
		}
		err := member.Validate()
		assert.Equal(t, exception.ErrInvalidLeaderType, err)
	})

	t.Run("fails with negative Point", func(t *testing.T) {
		member := &domain.ContestMember{
			UserID:     1,
			ContestID:  100,
			MemberType: domain.MemberTypeNormal,
			LeaderType: domain.LeaderTypeMember,
			Point:      -10,
		}
		err := member.Validate()
		assert.Equal(t, exception.ErrInvalidPoint, err)
	})
}

func TestContestMember_LeaderCreation(t *testing.T) {
	t.Run("leader has STAFF member type and LEADER leader type", func(t *testing.T) {
		member := domain.NewContestMemberAsLeader(1, 100)

		assert.True(t, member.IsLeader())
		assert.True(t, member.IsStaff())
	})

	t.Run("regular member is not leader and not staff", func(t *testing.T) {
		member := domain.NewContestMember(2, 100, domain.MemberTypeNormal, domain.LeaderTypeMember)

		assert.False(t, member.IsLeader())
		assert.False(t, member.IsStaff())
	})

	t.Run("staff member is not leader but is staff", func(t *testing.T) {
		member := domain.NewContestMember(3, 100, domain.MemberTypeStaff, domain.LeaderTypeMember)

		assert.False(t, member.IsLeader())
		assert.True(t, member.IsStaff())
	})
}
