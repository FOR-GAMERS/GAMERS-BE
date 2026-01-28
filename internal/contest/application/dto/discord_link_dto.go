package dto

// DiscordLinkRequiredResponse is returned when a user tries to perform an action
// that requires Discord account linking but hasn't linked their account yet.
type DiscordLinkRequiredResponse struct {
	Message  string `json:"message"`
	OAuthURL string `json:"oauth_url"`
}

// NewDiscordLinkRequiredResponse creates a new DiscordLinkRequiredResponse
func NewDiscordLinkRequiredResponse(message string) *DiscordLinkRequiredResponse {
	return &DiscordLinkRequiredResponse{
		Message:  message,
		OAuthURL: "/api/oauth2/discord/login",
	}
}
