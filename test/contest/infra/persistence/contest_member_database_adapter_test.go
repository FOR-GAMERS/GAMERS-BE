package persistence_test

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/contest/domain"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/contest/infra/persistence/adapter"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/exception"
	"github.com/FOR-GAMERS/GAMERS-BE/test/global/support"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type ContestMemberDatabaseAdapterTestSuite struct {
	suite.Suite
	container     *support.MySQLContainer
	db            *gorm.DB
	memberAdapter *adapter.ContestMemberDatabaseAdapter
	contestAdapter *adapter.ContestDatabaseAdapter
}

func (s *ContestMemberDatabaseAdapterTestSuite) SetupSuite() {
	ctx := context.Background()
	var err error

	s.container, err = support.SetupMySQLContainer(ctx)
	s.Require().NoError(err, "Failed to setup MySQL container")

	s.db = s.container.GetDB()

	// Auto-migrate schemas
	err = s.db.AutoMigrate(&domain.Contest{}, &domain.ContestMember{})
	s.Require().NoError(err, "Failed to migrate schemas")
}

func (s *ContestMemberDatabaseAdapterTestSuite) TearDownSuite() {
	ctx := context.Background()
	if s.container != nil {
		s.container.Teardown(ctx)
	}
}

func (s *ContestMemberDatabaseAdapterTestSuite) SetupTest() {
	// Clean up tables before each test (order matters for foreign keys)
	s.db.Exec("DELETE FROM contests_members")
	s.db.Exec("DELETE FROM contests")

	// Initialize adapters
	s.memberAdapter = adapter.NewContestMemberDatabaseAdapter(s.db)
	s.contestAdapter = adapter.NewContestDatabaseAdapter(s.db)
}

func (s *ContestMemberDatabaseAdapterTestSuite) createTestContest(title string) *domain.Contest {
	contest := &domain.Contest{
		Title:           title,
		Description:     "Test description",
		MaxTeamCount:    8,
		TotalPoint:      100,
		ContestType:     domain.ContestTypeTournament,
		ContestStatus:   domain.ContestStatusPending,
		StartedAt:       time.Now().Add(-1 * time.Hour),
		EndedAt:         time.Now().Add(48 * time.Hour),
		TotalTeamMember: 5,
	}
	saved, _ := s.contestAdapter.Save(contest)
	return saved
}

// ==================== Save Tests ====================

func (s *ContestMemberDatabaseAdapterTestSuite) TestSave_Success() {
	// Given
	contest := s.createTestContest("Test Contest")
	member := domain.NewContestMemberAsLeader(1, contest.ContestID)

	// When
	err := s.memberAdapter.Save(member)

	// Then
	s.NoError(err)
}

func (s *ContestMemberDatabaseAdapterTestSuite) TestSave_NormalMember() {
	// Given
	contest := s.createTestContest("Test Contest")
	member := domain.NewContestMember(2, contest.ContestID, domain.MemberTypeNormal, domain.LeaderTypeMember)

	// When
	err := s.memberAdapter.Save(member)

	// Then
	s.NoError(err)
}

func (s *ContestMemberDatabaseAdapterTestSuite) TestSave_InvalidMember() {
	// Given: Member with invalid user ID
	contest := s.createTestContest("Test Contest")
	member := &domain.ContestMember{
		UserID:     0, // Invalid
		ContestID:  contest.ContestID,
		MemberType: domain.MemberTypeNormal,
		LeaderType: domain.LeaderTypeMember,
	}

	// When
	err := s.memberAdapter.Save(member)

	// Then
	s.Error(err)
	s.Equal(exception.ErrInvalidUserID, err)
}

func (s *ContestMemberDatabaseAdapterTestSuite) TestSave_InvalidMemberType() {
	// Given
	contest := s.createTestContest("Test Contest")
	member := &domain.ContestMember{
		UserID:     1,
		ContestID:  contest.ContestID,
		MemberType: domain.MemberType("INVALID"),
		LeaderType: domain.LeaderTypeMember,
	}

	// When
	err := s.memberAdapter.Save(member)

	// Then
	s.Error(err)
	s.Equal(exception.ErrInvalidMemberType, err)
}

// ==================== GetByContestAndUser Tests ====================

func (s *ContestMemberDatabaseAdapterTestSuite) TestGetByContestAndUser_Success() {
	// Given
	contest := s.createTestContest("Test Contest")
	member := domain.NewContestMemberAsLeader(1, contest.ContestID)
	err := s.memberAdapter.Save(member)
	s.NoError(err)

	// When
	found, err := s.memberAdapter.GetByContestAndUser(contest.ContestID, 1)

	// Then
	s.NoError(err)
	s.NotNil(found)
	s.Equal(int64(1), found.UserID)
	s.Equal(contest.ContestID, found.ContestID)
	s.True(found.IsLeader())
	s.True(found.IsStaff())
}

func (s *ContestMemberDatabaseAdapterTestSuite) TestGetByContestAndUser_NotFound() {
	// Given
	contest := s.createTestContest("Test Contest")

	// When
	found, err := s.memberAdapter.GetByContestAndUser(contest.ContestID, 999)

	// Then
	s.Error(err)
	s.Nil(found)
	s.Equal(exception.ErrContestMemberNotFound, err)
}

// ==================== GetMembersByContest Tests ====================

func (s *ContestMemberDatabaseAdapterTestSuite) TestGetMembersByContest_Success() {
	// Given
	contest := s.createTestContest("Test Contest")

	leader := domain.NewContestMemberAsLeader(1, contest.ContestID)
	s.memberAdapter.Save(leader)

	member1 := domain.NewContestMember(2, contest.ContestID, domain.MemberTypeNormal, domain.LeaderTypeMember)
	s.memberAdapter.Save(member1)

	member2 := domain.NewContestMember(3, contest.ContestID, domain.MemberTypeStaff, domain.LeaderTypeMember)
	s.memberAdapter.Save(member2)

	// When
	members, err := s.memberAdapter.GetMembersByContest(contest.ContestID)

	// Then
	s.NoError(err)
	s.Len(members, 3)
}

func (s *ContestMemberDatabaseAdapterTestSuite) TestGetMembersByContest_Empty() {
	// Given
	contest := s.createTestContest("Empty Contest")

	// When
	members, err := s.memberAdapter.GetMembersByContest(contest.ContestID)

	// Then
	s.NoError(err)
	s.Empty(members)
}

// ==================== DeleteById Tests ====================

func (s *ContestMemberDatabaseAdapterTestSuite) TestDeleteById_Success() {
	// Given
	contest := s.createTestContest("Test Contest")
	member := domain.NewContestMember(10, contest.ContestID, domain.MemberTypeNormal, domain.LeaderTypeMember)
	s.memberAdapter.Save(member)

	// When
	err := s.memberAdapter.DeleteById(contest.ContestID, 10)

	// Then
	s.NoError(err)

	// Verify deletion
	_, err = s.memberAdapter.GetByContestAndUser(contest.ContestID, 10)
	s.Equal(exception.ErrContestMemberNotFound, err)
}

func (s *ContestMemberDatabaseAdapterTestSuite) TestDeleteById_NotFound() {
	// Given
	contest := s.createTestContest("Test Contest")

	// When
	err := s.memberAdapter.DeleteById(contest.ContestID, 999)

	// Then
	s.Error(err)
	s.Equal(exception.ErrContestMemberNotFound, err)
}

// ==================== SaveBatch Tests ====================

func (s *ContestMemberDatabaseAdapterTestSuite) TestSaveBatch_Success() {
	// Given
	contest := s.createTestContest("Batch Contest")
	members := []*domain.ContestMember{
		domain.NewContestMember(100, contest.ContestID, domain.MemberTypeNormal, domain.LeaderTypeMember),
		domain.NewContestMember(101, contest.ContestID, domain.MemberTypeNormal, domain.LeaderTypeMember),
		domain.NewContestMember(102, contest.ContestID, domain.MemberTypeNormal, domain.LeaderTypeMember),
	}

	// When
	err := s.memberAdapter.SaveBatch(members)

	// Then
	s.NoError(err)

	// Verify all members saved
	allMembers, err := s.memberAdapter.GetMembersByContest(contest.ContestID)
	s.NoError(err)
	s.Len(allMembers, 3)
}

func (s *ContestMemberDatabaseAdapterTestSuite) TestSaveBatch_Empty() {
	// When
	err := s.memberAdapter.SaveBatch([]*domain.ContestMember{})

	// Then
	s.NoError(err)
}

func (s *ContestMemberDatabaseAdapterTestSuite) TestSaveBatch_RollbackOnError() {
	// Given
	contest := s.createTestContest("Batch Contest")
	members := []*domain.ContestMember{
		domain.NewContestMember(200, contest.ContestID, domain.MemberTypeNormal, domain.LeaderTypeMember),
		{UserID: 0, ContestID: contest.ContestID, MemberType: domain.MemberTypeNormal, LeaderType: domain.LeaderTypeMember}, // Invalid
		domain.NewContestMember(202, contest.ContestID, domain.MemberTypeNormal, domain.LeaderTypeMember),
	}

	// When
	err := s.memberAdapter.SaveBatch(members)

	// Then
	s.Error(err)

	// Verify rollback - first member should not exist
	allMembers, _ := s.memberAdapter.GetMembersByContest(contest.ContestID)
	s.Empty(allMembers)
}

// ==================== UpdateMemberType Tests ====================

func (s *ContestMemberDatabaseAdapterTestSuite) TestUpdateMemberType_Success() {
	// Given
	contest := s.createTestContest("Update Type Contest")
	member := domain.NewContestMember(50, contest.ContestID, domain.MemberTypeNormal, domain.LeaderTypeMember)
	s.memberAdapter.Save(member)

	// When
	err := s.memberAdapter.UpdateMemberType(contest.ContestID, 50, domain.MemberTypeStaff)

	// Then
	s.NoError(err)

	// Verify update
	found, _ := s.memberAdapter.GetByContestAndUser(contest.ContestID, 50)
	s.Equal(domain.MemberTypeStaff, found.MemberType)
}

func (s *ContestMemberDatabaseAdapterTestSuite) TestUpdateMemberType_NotFound() {
	// Given
	contest := s.createTestContest("Update Type Contest")

	// When
	err := s.memberAdapter.UpdateMemberType(contest.ContestID, 999, domain.MemberTypeStaff)

	// Then
	s.Error(err)
	s.Equal(exception.ErrContestMemberNotFound, err)
}

func (s *ContestMemberDatabaseAdapterTestSuite) TestUpdateMemberType_DemoteStaff() {
	// Given
	contest := s.createTestContest("Demote Contest")
	member := domain.NewContestMember(60, contest.ContestID, domain.MemberTypeStaff, domain.LeaderTypeMember)
	s.memberAdapter.Save(member)

	// When
	err := s.memberAdapter.UpdateMemberType(contest.ContestID, 60, domain.MemberTypeNormal)

	// Then
	s.NoError(err)

	// Verify
	found, _ := s.memberAdapter.GetByContestAndUser(contest.ContestID, 60)
	s.Equal(domain.MemberTypeNormal, found.MemberType)
}

// ==================== Multiple Contests Tests ====================

func (s *ContestMemberDatabaseAdapterTestSuite) TestMember_InMultipleContests() {
	// Given: Same user in multiple contests
	contest1 := s.createTestContest("Contest 1")
	contest2 := s.createTestContest("Contest 2")

	userID := int64(99)

	member1 := domain.NewContestMemberAsLeader(userID, contest1.ContestID)
	member2 := domain.NewContestMember(userID, contest2.ContestID, domain.MemberTypeNormal, domain.LeaderTypeMember)

	// When
	err1 := s.memberAdapter.Save(member1)
	err2 := s.memberAdapter.Save(member2)

	// Then
	s.NoError(err1)
	s.NoError(err2)

	// Verify different roles in different contests
	found1, _ := s.memberAdapter.GetByContestAndUser(contest1.ContestID, userID)
	found2, _ := s.memberAdapter.GetByContestAndUser(contest2.ContestID, userID)

	s.True(found1.IsLeader())
	s.False(found2.IsLeader())
}

// ==================== Run the test suite ====================

func TestContestMemberDatabaseAdapterSuite(t *testing.T) {
	suite.Run(t, new(ContestMemberDatabaseAdapterTestSuite))
}
