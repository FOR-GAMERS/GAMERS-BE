package port

// NotificationHandlerPort defines the interface for sending notifications
// This interface is used by other services to send notifications
type NotificationHandlerPort interface {
	// Team invite notifications
	HandleTeamInviteReceived(inviteeUserID int64, inviterUsername, teamName string, gameID, contestID int64) error
	HandleTeamInviteAccepted(inviterUserID int64, inviteeUsername, teamName string, gameID, contestID int64) error
	HandleTeamInviteRejected(inviterUserID int64, inviteeUsername, teamName string, gameID, contestID int64) error

	// Contest application notifications
	HandleApplicationAccepted(userID, contestID int64, contestTitle string) error
	HandleApplicationRejected(userID, contestID int64, contestTitle, reason string) error
}
