package discord

import (
	"GAMERS-BE/internal/oauth2/application/dto"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

const (
	APIBaseURL       = "https://discord.com/api/v10"
	UserInfoEndpoint = "/users/@me"
)

type Client struct {
	httpClient *http.Client
}

func NewDiscordClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) GetUserInfo(token *oauth2.Token) (*dto.DiscordUserInfo, error) {
	req, err := http.NewRequest("GET", APIBaseURL+UserInfoEndpoint, nil)
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
