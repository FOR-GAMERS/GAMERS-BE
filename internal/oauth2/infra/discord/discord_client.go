package discord

import (
	"GAMERS-BE/internal/oauth2/application/dto"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"golang.org/x/oauth2"
)

const (
	DiscordAPIBaseURL = "https://discord.com/api/v10"
	UserInfoEndpoint  = "/users/@me"
)

type DiscordClient struct {
	httpClient *http.Client
}

func NewDiscordClient() *DiscordClient {
	return &DiscordClient{
		httpClient: &http.Client{},
	}
}

func (c *DiscordClient) GetUserInfo(token *oauth2.Token) (*dto.DiscordUserInfo, error) {
	req, err := http.NewRequest("GET", DiscordAPIBaseURL+UserInfoEndpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, errors.New("failed to get user info from Discord: " + string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo dto.DiscordUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}
