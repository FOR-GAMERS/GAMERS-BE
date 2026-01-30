package persistence_test

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/contest/domain"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/contest/infra/persistence/adapter"
	commonDto "github.com/FOR-GAMERS/GAMERS-BE/internal/global/common/dto"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/exception"
	"github.com/FOR-GAMERS/GAMERS-BE/test/global/support"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type ContestDatabaseAdapterTestSuite struct {
	suite.Suite
	container *support.MySQLContainer
	db        *gorm.DB
	adapter   *adapter.ContestDatabaseAdapter
}

func (s *ContestDatabaseAdapterTestSuite) SetupSuite() {
	ctx := context.Background()
	var err error

	s.container, err = support.SetupMySQLContainer(ctx)
	s.Require().NoError(err, "Failed to setup MySQL container")

	s.db = s.container.GetDB()

	// Auto-migrate schema
	err = s.db.AutoMigrate(&domain.Contest{})
	s.Require().NoError(err, "Failed to migrate schema")
}

func (s *ContestDatabaseAdapterTestSuite) TearDownSuite() {
	ctx := context.Background()
	if s.container != nil {
		s.container.Teardown(ctx)
	}
}

func (s *ContestDatabaseAdapterTestSuite) SetupTest() {
	// Clean up table before each test
	s.db.Exec("DELETE FROM contests")

	// Initialize adapter
	s.adapter = adapter.NewContestDatabaseAdapter(s.db)
}

func (s *ContestDatabaseAdapterTestSuite) createTestContest(title string) *domain.Contest {
	return &domain.Contest{
		Title:           title,
		Description:     "Test description",
		MaxTeamCount:    8,
		TotalPoint:      100,
		ContestType:     domain.ContestTypeTournament,
		ContestStatus:   domain.ContestStatusPending,
		StartedAt:       time.Now().Add(-1 * time.Hour),
		EndedAt:         time.Now().Add(48 * time.Hour),
		TotalTeamMember: 5,
		AutoStart:       false,
	}
}

// ==================== Save Tests ====================

func (s *ContestDatabaseAdapterTestSuite) TestSave_Success() {
	// Given
	contest := s.createTestContest("Test Tournament")

	// When
	saved, err := s.adapter.Save(contest)

	// Then
	s.NoError(err)
	s.NotNil(saved)
	s.NotZero(saved.ContestID)
	s.Equal("Test Tournament", saved.Title)
}

func (s *ContestDatabaseAdapterTestSuite) TestSave_AllFields() {
	// Given
	guildId := "guild123"
	channelId := "channel456"
	thumbnail := "thumbnail.jpg"
	bannerKey := "banner.jpg"

	contest := &domain.Contest{
		Title:                "Full Contest",
		Description:          "Full description",
		MaxTeamCount:         16,
		TotalPoint:           200,
		ContestType:          domain.ContestTypeLeague,
		ContestStatus:        domain.ContestStatusPending,
		StartedAt:            time.Now().Add(1 * time.Hour),
		EndedAt:              time.Now().Add(72 * time.Hour),
		AutoStart:            true,
		TotalTeamMember:      10,
		DiscordGuildId:       &guildId,
		DiscordTextChannelId: &channelId,
		Thumbnail:            &thumbnail,
		BannerKey:            &bannerKey,
	}

	// When
	saved, err := s.adapter.Save(contest)

	// Then
	s.NoError(err)
	s.NotNil(saved)
	s.Equal("Full Contest", saved.Title)
	s.Equal(domain.ContestTypeLeague, saved.ContestType)
	s.True(saved.AutoStart)
	s.NotNil(saved.DiscordGuildId)
	s.Equal("guild123", *saved.DiscordGuildId)
}

func (s *ContestDatabaseAdapterTestSuite) TestSave_UpdateExisting() {
	// Given
	contest := s.createTestContest("Original Title")
	saved, err := s.adapter.Save(contest)
	s.NoError(err)

	// When
	saved.Title = "Updated Title"
	saved.MaxTeamCount = 16
	updated, err := s.adapter.Save(saved)

	// Then
	s.NoError(err)
	s.Equal("Updated Title", updated.Title)
	s.Equal(16, updated.MaxTeamCount)
	s.Equal(saved.ContestID, updated.ContestID)
}

// ==================== GetContestById Tests ====================

func (s *ContestDatabaseAdapterTestSuite) TestGetContestById_Success() {
	// Given
	contest := s.createTestContest("Find Me")
	saved, err := s.adapter.Save(contest)
	s.NoError(err)

	// When
	found, err := s.adapter.GetContestById(saved.ContestID)

	// Then
	s.NoError(err)
	s.NotNil(found)
	s.Equal(saved.ContestID, found.ContestID)
	s.Equal("Find Me", found.Title)
}

func (s *ContestDatabaseAdapterTestSuite) TestGetContestById_NotFound() {
	// When
	found, err := s.adapter.GetContestById(99999)

	// Then
	s.Error(err)
	s.Nil(found)
	s.Equal(exception.ErrContestNotFound, err)
}

// ==================== GetContests Tests ====================

func (s *ContestDatabaseAdapterTestSuite) TestGetContests_EmptyResult() {
	// When
	contests, total, err := s.adapter.GetContests(0, 10, nil, nil)

	// Then
	s.NoError(err)
	s.Empty(contests)
	s.Equal(int64(0), total)
}

func (s *ContestDatabaseAdapterTestSuite) TestGetContests_WithPagination() {
	// Given: Create 5 contests
	for i := 1; i <= 5; i++ {
		contest := s.createTestContest("Contest " + string(rune('A'-1+i)))
		_, err := s.adapter.Save(contest)
		s.NoError(err)
	}

	// When: Get first page
	contests, total, err := s.adapter.GetContests(0, 2, nil, nil)

	// Then
	s.NoError(err)
	s.Len(contests, 2)
	s.Equal(int64(5), total)

	// When: Get second page
	contests, total, err = s.adapter.GetContests(2, 2, nil, nil)

	// Then
	s.NoError(err)
	s.Len(contests, 2)
	s.Equal(int64(5), total)
}

func (s *ContestDatabaseAdapterTestSuite) TestGetContests_WithSorting() {
	// Given: Create contests with different names
	names := []string{"Zulu", "Alpha", "Mike"}
	for _, name := range names {
		contest := s.createTestContest(name)
		_, err := s.adapter.Save(contest)
		s.NoError(err)
		time.Sleep(10 * time.Millisecond) // Ensure different created_at
	}

	// When: Sort by title ASC
	sortReq := commonDto.NewSortRequest("title", "asc", []string{"title", "created_at"})
	contests, _, err := s.adapter.GetContests(0, 10, sortReq, nil)

	// Then
	s.NoError(err)
	s.Len(contests, 3)
	s.Equal("Alpha", contests[0].Title)
	s.Equal("Mike", contests[1].Title)
	s.Equal("Zulu", contests[2].Title)
}

func (s *ContestDatabaseAdapterTestSuite) TestGetContests_WithTitleSearch() {
	// Given
	_, _ = s.adapter.Save(s.createTestContest("Valorant Tournament"))
	_, _ = s.adapter.Save(s.createTestContest("Valorant League"))
	_, _ = s.adapter.Save(s.createTestContest("LoL Championship"))

	// When
	searchTitle := "Valorant"
	contests, total, err := s.adapter.GetContests(0, 10, nil, &searchTitle)

	// Then
	s.NoError(err)
	s.Len(contests, 2)
	s.Equal(int64(2), total)
}

func (s *ContestDatabaseAdapterTestSuite) TestGetContests_WithPartialTitleSearch() {
	// Given
	_, _ = s.adapter.Save(s.createTestContest("Spring Tournament 2024"))
	_, _ = s.adapter.Save(s.createTestContest("Summer Tournament 2024"))
	_, _ = s.adapter.Save(s.createTestContest("Winter League"))

	// When
	searchTitle := "Tournament"
	contests, total, err := s.adapter.GetContests(0, 10, nil, &searchTitle)

	// Then
	s.NoError(err)
	s.Len(contests, 2)
	s.Equal(int64(2), total)
}

// ==================== DeleteContestById Tests ====================

func (s *ContestDatabaseAdapterTestSuite) TestDeleteContestById_Success() {
	// Given
	contest := s.createTestContest("To Be Deleted")
	saved, err := s.adapter.Save(contest)
	s.NoError(err)

	// When
	err = s.adapter.DeleteContestById(saved.ContestID)

	// Then
	s.NoError(err)

	// Verify deletion
	_, err = s.adapter.GetContestById(saved.ContestID)
	s.Equal(exception.ErrContestNotFound, err)
}

func (s *ContestDatabaseAdapterTestSuite) TestDeleteContestById_NonExistent() {
	// When
	err := s.adapter.DeleteContestById(99999)

	// Then: Should not return error (GORM behavior)
	s.NoError(err)
}

// ==================== UpdateContest Tests ====================

func (s *ContestDatabaseAdapterTestSuite) TestUpdateContest_Success() {
	// Given
	contest := s.createTestContest("Original")
	saved, err := s.adapter.Save(contest)
	s.NoError(err)

	// When
	saved.Title = "Updated"
	saved.ContestStatus = domain.ContestStatusActive
	err = s.adapter.UpdateContest(saved)

	// Then
	s.NoError(err)

	// Verify update
	found, err := s.adapter.GetContestById(saved.ContestID)
	s.NoError(err)
	s.Equal("Updated", found.Title)
	s.Equal(domain.ContestStatusActive, found.ContestStatus)
}

func (s *ContestDatabaseAdapterTestSuite) TestUpdateContest_StatusTransition() {
	// Given
	contest := s.createTestContest("Status Test")
	saved, err := s.adapter.Save(contest)
	s.NoError(err)
	s.Equal(domain.ContestStatusPending, saved.ContestStatus)

	// When: Transition to ACTIVE
	saved.ContestStatus = domain.ContestStatusActive
	err = s.adapter.UpdateContest(saved)
	s.NoError(err)

	// Then
	found, _ := s.adapter.GetContestById(saved.ContestID)
	s.Equal(domain.ContestStatusActive, found.ContestStatus)

	// When: Transition to FINISHED
	found.ContestStatus = domain.ContestStatusFinished
	err = s.adapter.UpdateContest(found)
	s.NoError(err)

	// Then
	final, _ := s.adapter.GetContestById(saved.ContestID)
	s.Equal(domain.ContestStatusFinished, final.ContestStatus)
}

// ==================== Run the test suite ====================

func TestContestDatabaseAdapterSuite(t *testing.T) {
	suite.Run(t, new(ContestDatabaseAdapterTestSuite))
}
