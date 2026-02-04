package presentation

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/common/router"
	_ "github.com/FOR-GAMERS/GAMERS-BE/internal/global/response" // for swagger
	"github.com/FOR-GAMERS/GAMERS-BE/internal/oauth2/application"
	"github.com/FOR-GAMERS/GAMERS-BE/internal/oauth2/application/dto"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type DiscordController struct {
	router        *router.Router
	oauth2Service *application.DiscordService
	webURL        string
	cookieDomain  string
}

func NewDiscordController(router *router.Router, oauth2Service *application.DiscordService, webURL string, cookieDomain string) *DiscordController {
	return &DiscordController{
		router:        router,
		oauth2Service: oauth2Service,
		webURL:        webURL,
		cookieDomain:  cookieDomain,
	}
}

func (c *DiscordController) RegisterRoutes() {
	oauth2Group := c.router.PublicGroup("/api/oauth2")
	{
		oauth2Group.GET("/discord/login", c.DiscordLogin)
		oauth2Group.GET("/discord/callback", c.DiscordCallback)
	}
}

// DiscordLogin godoc
// @Summary Discord OAuth2 Login
// @Description Redirect to Discord OAuth2 login page
// @Tags oauth2
// @Accept json
// @Produce json
// @Success 302 {string} string "Redirect to Discord login page"
// @Router /api/oauth2/discord/login [get]
func (c *DiscordController) DiscordLogin(ctx *gin.Context) {
	loginURL, err := c.oauth2Service.GetDiscordLoginURL()

	if err != nil {
		ctx.Error(err)
		return
	}

	// Redirect to Discord login page
	ctx.Redirect(302, loginURL)
}

// DiscordCallback godoc
// @Summary Discord OAuth2 Callback
// @Description Handle Discord OAuth2 callback
// @Tags oauth2
// @Accept json
// @Produce json
// @Param code query string true "Authorization code"
// @Param state query string false "State"
// @Success 302 {string} string "Redirect to frontend with cookies"
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/oauth2/discord/callback [get]
func (c *DiscordController) DiscordCallback(ctx *gin.Context) {
	var req dto.DiscordCallbackRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.Error(err)
		return
	}

	result, err := c.oauth2Service.HandleDiscordCallback(&req)
	if err != nil {
		ctx.Error(err)
		return
	}

	// Determine if we're in production (HTTPS) or development (HTTP)
	isSecure := os.Getenv("ENV") == "production" || strings.HasPrefix(c.webURL, "https")
	sameSite := http.SameSiteLaxMode
	if isSecure {
		sameSite = http.SameSiteNoneMode
	}

	// Set access token cookie
	ctx.SetSameSite(sameSite)
	ctx.SetCookie(
		"access_token",
		result.AccessToken,
		60*15, // 15 minutes (match JWT_ACCESS_DURATION)
		"/",
		c.cookieDomain, // e.g. ".gamers.io.kr" for cross-subdomain sharing
		isSecure,
		true, // HttpOnly
	)

	// Set refresh token cookie
	ctx.SetCookie(
		"refresh_token",
		result.RefreshToken,
		60*60*24*7, // 7 days (match JWT_REFRESH_DURATION)
		"/",
		c.cookieDomain,
		isSecure,
		true, // HttpOnly
	)

	// Set is_new_user cookie (not HttpOnly so frontend can read it)
	isNewUserValue := "false"
	if result.IsNewUser {
		isNewUserValue = "true"
	}
	ctx.SetCookie(
		"is_new_user",
		isNewUserValue,
		60*5, // 5 minutes (short-lived, just for frontend to check)
		"/",
		c.cookieDomain,
		isSecure,
		false, // Not HttpOnly so frontend can read
	)

	// Redirect to frontend login success page
	redirectURL := c.webURL + "/login/success"
	ctx.Redirect(http.StatusFound, redirectURL)
}
