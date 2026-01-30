package application

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/game/application/port"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/game/domain"
	"context"
	"fmt"
	"log"
)

// TeamPersistenceHandler handles DB persistence for team events
type TeamPersistenceHandler struct {
	teamDBRepository port.TeamDatabasePort
}

// NewTeamPersistenceHandler creates a new persistence handler
func NewTeamPersistenceHandler(teamDBRepository port.TeamDatabasePort) *TeamPersistenceHandler {
	return &TeamPersistenceHandler{
		teamDBRepository: teamDBRepository,
	}
}

// HandleTeamPersistence handles team persistence events
func (h *TeamPersistenceHandler) HandleTeamPersistence(ctx context.Context, event *port.TeamPersistenceEvent) error {
	log.Printf("Handling team persistence event: type=%s, contestID=%d, teamID=%d",
		event.EventType, event.ContestID, event.TeamID)

	switch event.EventType {
	case port.TeamPersistenceCreated:
		return h.handleTeamCreated(ctx, event)
	case port.TeamPersistenceMemberAdded:
		return h.handleMemberAdded(ctx, event)
	case port.TeamPersistenceMemberRemoved:
		return h.handleMemberRemoved(ctx, event)
	case port.TeamPersistenceFinalized:
		return h.handleTeamFinalized(ctx, event)
	case port.TeamPersistenceDeleted:
		return h.handleTeamDeleted(ctx, event)
	default:
		return fmt.Errorf("unknown event type: %s", event.EventType)
	}
}

// handleTeamCreated persists a newly created team to DB
func (h *TeamPersistenceHandler) handleTeamCreated(ctx context.Context, event *port.TeamPersistenceEvent) error {
	teamName := ""
	if event.TeamName != nil {
		teamName = *event.TeamName
	}

	// Create team in DB
	team := domain.NewTeam(event.ContestID, teamName)
	savedTeam, err := h.teamDBRepository.Save(team)
	if err != nil {
		return fmt.Errorf("failed to save team: %w", err)
	}

	// Save members if present
	if len(event.Members) > 0 {
		for _, member := range event.Members {
			memberType := domain.TeamMemberTypeMember
			if member.MemberType == port.TeamMemberTypeLeader {
				memberType = domain.TeamMemberTypeLeader
			}

			dbMember := domain.NewTeamMember(savedTeam.TeamID, member.UserID, memberType)
			dbMember.JoinedAt = member.JoinedAt

			if _, err := h.teamDBRepository.SaveMember(dbMember); err != nil {
				return fmt.Errorf("failed to save member %d: %w", member.UserID, err)
			}
		}
	}

	log.Printf("Successfully persisted team created event: teamID=%d, contestID=%d",
		savedTeam.TeamID, event.ContestID)

	return nil
}

// handleMemberAdded persists a new member addition to DB
func (h *TeamPersistenceHandler) handleMemberAdded(ctx context.Context, event *port.TeamPersistenceEvent) error {
	if event.MemberUserID == nil {
		return fmt.Errorf("member_user_id is required for member_added event")
	}

	// Find the team by contest ID
	teams, err := h.teamDBRepository.GetByContestID(event.ContestID)
	if err != nil {
		return fmt.Errorf("failed to get teams for contest %d: %w", event.ContestID, err)
	}

	if len(teams) == 0 {
		return fmt.Errorf("no team found for contest %d", event.ContestID)
	}

	// Get the team (assuming one team per contest for now, or match by team name if needed)
	var targetTeam *domain.Team
	if event.TeamName != nil {
		for _, t := range teams {
			if t.TeamName == *event.TeamName {
				targetTeam = t
				break
			}
		}
	}
	if targetTeam == nil {
		targetTeam = teams[0] // Fallback to first team
	}

	memberType := domain.TeamMemberTypeMember
	if event.MemberType != nil && *event.MemberType == port.TeamMemberTypeLeader {
		memberType = domain.TeamMemberTypeLeader
	}

	dbMember := domain.NewTeamMember(targetTeam.TeamID, *event.MemberUserID, memberType)
	if event.MemberJoinedAt != nil {
		dbMember.JoinedAt = *event.MemberJoinedAt
	}

	if _, err := h.teamDBRepository.SaveMember(dbMember); err != nil {
		return fmt.Errorf("failed to save member: %w", err)
	}

	log.Printf("Successfully persisted member added event: teamID=%d, userID=%d",
		targetTeam.TeamID, *event.MemberUserID)

	return nil
}

// handleMemberRemoved persists a member removal from DB
func (h *TeamPersistenceHandler) handleMemberRemoved(ctx context.Context, event *port.TeamPersistenceEvent) error {
	if event.MemberUserID == nil {
		return fmt.Errorf("member_user_id is required for member_removed event")
	}

	// Find the team by contest ID
	teams, err := h.teamDBRepository.GetByContestID(event.ContestID)
	if err != nil {
		return fmt.Errorf("failed to get teams for contest %d: %w", event.ContestID, err)
	}

	if len(teams) == 0 {
		// Team not persisted yet, nothing to remove
		log.Printf("No team found for contest %d, skipping member removal", event.ContestID)
		return nil
	}

	var targetTeam *domain.Team
	if event.TeamName != nil {
		for _, t := range teams {
			if t.TeamName == *event.TeamName {
				targetTeam = t
				break
			}
		}
	}
	if targetTeam == nil {
		targetTeam = teams[0]
	}

	if err := h.teamDBRepository.DeleteMemberByTeamAndUser(targetTeam.TeamID, *event.MemberUserID); err != nil {
		return fmt.Errorf("failed to delete member: %w", err)
	}

	log.Printf("Successfully persisted member removed event: teamID=%d, userID=%d",
		targetTeam.TeamID, *event.MemberUserID)

	return nil
}

// handleTeamFinalized persists all team data to DB
func (h *TeamPersistenceHandler) handleTeamFinalized(ctx context.Context, event *port.TeamPersistenceEvent) error {
	teamName := ""
	if event.TeamName != nil {
		teamName = *event.TeamName
	}

	// Check if team already exists
	existingTeam, err := h.teamDBRepository.GetByContestAndName(event.ContestID, teamName)
	if err == nil && existingTeam != nil {
		// Team already exists, update members
		log.Printf("Team already exists for contest %d, updating members", event.ContestID)

		// Delete existing members and re-add
		if err := h.teamDBRepository.DeleteAllMembersByTeamID(existingTeam.TeamID); err != nil {
			return fmt.Errorf("failed to delete existing members: %w", err)
		}

		// Add all members
		for _, member := range event.Members {
			memberType := domain.TeamMemberTypeMember
			if member.MemberType == port.TeamMemberTypeLeader {
				memberType = domain.TeamMemberTypeLeader
			}

			dbMember := domain.NewTeamMember(existingTeam.TeamID, member.UserID, memberType)
			dbMember.JoinedAt = member.JoinedAt

			if _, err := h.teamDBRepository.SaveMember(dbMember); err != nil {
				return fmt.Errorf("failed to save member %d: %w", member.UserID, err)
			}
		}

		log.Printf("Successfully updated finalized team: teamID=%d", existingTeam.TeamID)
		return nil
	}

	// Create new team
	team := domain.NewTeam(event.ContestID, teamName)
	savedTeam, err := h.teamDBRepository.Save(team)
	if err != nil {
		return fmt.Errorf("failed to save team: %w", err)
	}

	// Save all members
	for _, member := range event.Members {
		memberType := domain.TeamMemberTypeMember
		if member.MemberType == port.TeamMemberTypeLeader {
			memberType = domain.TeamMemberTypeLeader
		}

		dbMember := domain.NewTeamMember(savedTeam.TeamID, member.UserID, memberType)
		dbMember.JoinedAt = member.JoinedAt

		if _, err := h.teamDBRepository.SaveMember(dbMember); err != nil {
			return fmt.Errorf("failed to save member %d: %w", member.UserID, err)
		}
	}

	log.Printf("Successfully persisted finalized team: teamID=%d, contestID=%d, members=%d",
		savedTeam.TeamID, event.ContestID, len(event.Members))

	return nil
}

// handleTeamDeleted removes team from DB
func (h *TeamPersistenceHandler) handleTeamDeleted(ctx context.Context, event *port.TeamPersistenceEvent) error {
	// Find the team by contest ID
	teams, err := h.teamDBRepository.GetByContestID(event.ContestID)
	if err != nil {
		return fmt.Errorf("failed to get teams for contest %d: %w", event.ContestID, err)
	}

	if len(teams) == 0 {
		// Team not persisted yet, nothing to delete
		log.Printf("No team found for contest %d, skipping deletion", event.ContestID)
		return nil
	}

	var targetTeam *domain.Team
	if event.TeamName != nil {
		for _, t := range teams {
			if t.TeamName == *event.TeamName {
				targetTeam = t
				break
			}
		}
	}
	if targetTeam == nil {
		targetTeam = teams[0]
	}

	// Delete all members first
	if err := h.teamDBRepository.DeleteAllMembersByTeamID(targetTeam.TeamID); err != nil {
		return fmt.Errorf("failed to delete team members: %w", err)
	}

	// Delete the team
	if err := h.teamDBRepository.Delete(targetTeam.TeamID); err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}

	log.Printf("Successfully persisted team deleted event: teamID=%d, contestID=%d",
		targetTeam.TeamID, event.ContestID)

	return nil
}
