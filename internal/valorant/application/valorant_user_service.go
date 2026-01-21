package application

import (
	"GAMERS-BE/internal/global/exception"
	pointDomain "GAMERS-BE/internal/point/domain"
	pointPort "GAMERS-BE/internal/point/application/port"
	userCommandPort "GAMERS-BE/internal/user/application/port/command"
	userQueryPort "GAMERS-BE/internal/user/application/port/port"
	"GAMERS-BE/internal/valorant/application/dto"
	"GAMERS-BE/internal/valorant/application/port"
	"math"
	"strings"
)

type ValorantUserService struct {
	valorantApi        port.ValorantApiPort
	userQueryPort      userQueryPort.UserQueryPort
	userCommandPort    userCommandPort.UserCommandPort
	scoreTablePort     pointPort.ValorantScoreTableDatabasePort
}

func NewValorantUserService(
	valorantApi port.ValorantApiPort,
	userQueryPort userQueryPort.UserQueryPort,
	userCommandPort userCommandPort.UserCommandPort,
	scoreTablePort pointPort.ValorantScoreTableDatabasePort,
) *ValorantUserService {
	return &ValorantUserService{
		valorantApi:     valorantApi,
		userQueryPort:   userQueryPort,
		userCommandPort: userCommandPort,
		scoreTablePort:  scoreTablePort,
	}
}

// RegisterValorant registers a Valorant account for a user
func (s *ValorantUserService) RegisterValorant(userId int64, req *dto.RegisterValorantRequest) (*dto.ValorantInfoResponse, error) {
	// Validate region
	if !isValidRegion(req.Region) {
		return nil, exception.ErrInvalidValorantRegion
	}

	// Get user
	user, err := s.userQueryPort.FindById(userId)
	if err != nil {
		return nil, err
	}

	// Check if Valorant is already linked
	if user.HasValorantLinked() {
		return nil, exception.ErrValorantAlreadyLinked
	}

	// Fetch MMR data from Valorant API
	mmrData, err := s.valorantApi.GetMMRByName(req.Region, req.RiotName, req.RiotTag)
	if err != nil {
		return nil, err
	}

	// Update user with Valorant info
	user.UpdateValorantInfo(
		req.RiotName,
		req.RiotTag,
		req.Region,
		mmrData.CurrentTier,
		mmrData.CurrentTierPatched,
		mmrData.Elo,
		mmrData.RankingInTier,
		mmrData.PeakTier,
		mmrData.PeakTierPatched,
	)

	if err := s.userCommandPort.UpdateValorantInfo(user); err != nil {
		return nil, err
	}

	return &dto.ValorantInfoResponse{
		RiotName:           req.RiotName,
		RiotTag:            req.RiotTag,
		Region:             req.Region,
		CurrentTier:        mmrData.CurrentTier,
		CurrentTierPatched: mmrData.CurrentTierPatched,
		Elo:                mmrData.Elo,
		RankingInTier:      mmrData.RankingInTier,
		PeakTier:           mmrData.PeakTier,
		PeakTierPatched:    mmrData.PeakTierPatched,
		UpdatedAt:          user.ValorantUpdatedAt,
		RefreshNeeded:      false,
	}, nil
}

// RefreshValorant refreshes Valorant data for a user
func (s *ValorantUserService) RefreshValorant(userId int64) (*dto.ValorantInfoResponse, error) {
	// Get user
	user, err := s.userQueryPort.FindById(userId)
	if err != nil {
		return nil, err
	}

	// Check if Valorant is linked
	if !user.HasValorantLinked() {
		return nil, exception.ErrValorantNotLinked
	}

	// Fetch MMR data from Valorant API
	mmrData, err := s.valorantApi.GetMMRByName(*user.Region, *user.RiotName, *user.RiotTag)
	if err != nil {
		return nil, err
	}

	// Keep peak tier if current peak is higher
	peakTier := mmrData.PeakTier
	peakTierPatched := mmrData.PeakTierPatched
	if user.PeakTier != nil && *user.PeakTier > mmrData.PeakTier {
		peakTier = *user.PeakTier
		peakTierPatched = *user.PeakTierPatched
	}

	// Update user with new Valorant info
	user.UpdateValorantInfo(
		*user.RiotName,
		*user.RiotTag,
		*user.Region,
		mmrData.CurrentTier,
		mmrData.CurrentTierPatched,
		mmrData.Elo,
		mmrData.RankingInTier,
		peakTier,
		peakTierPatched,
	)

	if err := s.userCommandPort.UpdateValorantInfo(user); err != nil {
		return nil, err
	}

	return &dto.ValorantInfoResponse{
		RiotName:           *user.RiotName,
		RiotTag:            *user.RiotTag,
		Region:             *user.Region,
		CurrentTier:        mmrData.CurrentTier,
		CurrentTierPatched: mmrData.CurrentTierPatched,
		Elo:                mmrData.Elo,
		RankingInTier:      mmrData.RankingInTier,
		PeakTier:           peakTier,
		PeakTierPatched:    peakTierPatched,
		UpdatedAt:          user.ValorantUpdatedAt,
		RefreshNeeded:      false,
	}, nil
}

// GetValorantInfo returns the stored Valorant information for a user
func (s *ValorantUserService) GetValorantInfo(userId int64) (*dto.ValorantInfoResponse, error) {
	// Get user
	user, err := s.userQueryPort.FindById(userId)
	if err != nil {
		return nil, err
	}

	// Check if Valorant is linked
	if !user.HasValorantLinked() {
		return nil, exception.ErrValorantNotLinked
	}

	return &dto.ValorantInfoResponse{
		RiotName:           *user.RiotName,
		RiotTag:            *user.RiotTag,
		Region:             *user.Region,
		CurrentTier:        *user.CurrentTier,
		CurrentTierPatched: *user.CurrentTierPatched,
		Elo:                *user.Elo,
		RankingInTier:      *user.RankingInTier,
		PeakTier:           *user.PeakTier,
		PeakTierPatched:    *user.PeakTierPatched,
		UpdatedAt:          user.ValorantUpdatedAt,
		RefreshNeeded:      user.IsValorantRefreshNeeded(),
	}, nil
}

// UnlinkValorant removes Valorant account link for a user
func (s *ValorantUserService) UnlinkValorant(userId int64) error {
	// Get user
	user, err := s.userQueryPort.FindById(userId)
	if err != nil {
		return err
	}

	// Check if Valorant is linked
	if !user.HasValorantLinked() {
		return exception.ErrValorantNotLinked
	}

	return s.userCommandPort.ClearValorantInfo(userId)
}

// CalculateContestPoint calculates the contest point for a user
func (s *ValorantUserService) CalculateContestPoint(userId int64, scoreTableId int64) (*dto.ContestPointResponse, error) {
	// Get user
	user, err := s.userQueryPort.FindById(userId)
	if err != nil {
		return nil, err
	}

	// Check if Valorant is linked
	if !user.HasValorantLinked() {
		return nil, exception.ErrValorantNotLinked
	}

	// Get score table
	scoreTable, err := s.scoreTablePort.GetByID(scoreTableId)
	if err != nil {
		return nil, err
	}

	// Calculate points
	currentTierPoint := getTierPoint(user.GetCurrentTierName(), scoreTable)
	peakTierPoint := getTierPoint(user.GetPeakTierName(), scoreTable)
	finalPoint := int(math.Round(float64(currentTierPoint+peakTierPoint) / 2))

	refreshNeeded := user.IsValorantRefreshNeeded()
	refreshMessage := ""
	if refreshNeeded {
		refreshMessage = "마지막 갱신 후 24시간이 지났습니다. 갱신이 필요합니다."
	}

	return &dto.ContestPointResponse{
		UserID:             userId,
		RiotName:           *user.RiotName,
		RiotTag:            *user.RiotTag,
		CurrentTierPatched: *user.CurrentTierPatched,
		CurrentTierPoint:   currentTierPoint,
		PeakTierPatched:    *user.PeakTierPatched,
		PeakTierPoint:      peakTierPoint,
		FinalPoint:         finalPoint,
		RefreshNeeded:      refreshNeeded,
		RefreshMessage:     refreshMessage,
	}, nil
}

func isValidRegion(region string) bool {
	validRegions := map[string]bool{
		"ap":    true,
		"br":    true,
		"eu":    true,
		"kr":    true,
		"latam": true,
		"na":    true,
	}
	return validRegions[strings.ToLower(region)]
}

func getTierPoint(tierName string, scoreTable *pointDomain.ValorantScoreTable) int {
	tierName = strings.ToLower(tierName)
	switch tierName {
	case "radiant":
		return scoreTable.Radiant
	case "immortal":
		return scoreTable.Immortal
	case "ascendant":
		return scoreTable.Ascendant
	case "diamond":
		return scoreTable.Diamond
	case "platinum":
		return scoreTable.Platinum
	case "gold":
		return scoreTable.Gold
	case "silver":
		return scoreTable.Silver
	case "bronze":
		return scoreTable.Bronze
	case "iron":
		return scoreTable.Iron
	default:
		return 0
	}
}
