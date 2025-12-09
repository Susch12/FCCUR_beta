package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	ErrOAuthStateMismatch = errors.New("oauth state mismatch")
	ErrOAuthExchange      = errors.New("oauth exchange failed")
)

// OAuth2Config holds OAuth2 configuration
type OAuth2Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
	Enabled      bool
}

// MicrosoftOAuth2Config creates Microsoft/Azure AD OAuth2 config
func NewMicrosoftOAuth2Config(clientID, clientSecret, redirectURL, tenant string) *OAuth2Config {
	if tenant == "" {
		tenant = "common" // Multi-tenant
	}

	return &OAuth2Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"openid", "profile", "email", "User.Read"},
		AuthURL:      fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/authorize", tenant),
		TokenURL:     fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenant),
		UserInfoURL:  "https://graph.microsoft.com/v1.0/me",
		Enabled:      clientID != "" && clientSecret != "",
	}
}

// GenerateState generates a random state parameter for OAuth2
func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GetAuthURL returns the OAuth2 authorization URL
func (c *OAuth2Config) GetAuthURL(state string) string {
	params := url.Values{}
	params.Add("client_id", c.ClientID)
	params.Add("response_type", "code")
	params.Add("redirect_uri", c.RedirectURL)
	params.Add("scope", strings.Join(c.Scopes, " "))
	params.Add("state", state)
	params.Add("response_mode", "query")

	return c.AuthURL + "?" + params.Encode()
}

// OAuth2Token represents an OAuth2 token response
type OAuth2Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
}

// ExchangeCode exchanges authorization code for access token
func (c *OAuth2Config) ExchangeCode(ctx context.Context, code string) (*OAuth2Token, error) {
	data := url.Values{}
	data.Set("client_id", c.ClientID)
	data.Set("client_secret", c.ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", c.RedirectURL)
	data.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, "POST", c.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: %s", ErrOAuthExchange, string(body))
	}

	var token OAuth2Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}

	return &token, nil
}

// MicrosoftUserInfo represents Microsoft Graph user information
type MicrosoftUserInfo struct {
	ID                string `json:"id"`
	DisplayName       string `json:"displayName"`
	GivenName         string `json:"givenName"`
	Surname           string `json:"surname"`
	UserPrincipalName string `json:"userPrincipalName"`
	Mail              string `json:"mail"`
}

// GetUserInfo retrieves user information using access token
func (c *OAuth2Config) GetUserInfo(ctx context.Context, accessToken string) (*MicrosoftUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info: %s", string(body))
	}

	var userInfo MicrosoftUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// GetEmail extracts email from Microsoft user info
func (u *MicrosoftUserInfo) GetEmail() string {
	if u.Mail != "" {
		return u.Mail
	}
	return u.UserPrincipalName
}

// GetFullName extracts full name from Microsoft user info
func (u *MicrosoftUserInfo) GetFullName() string {
	if u.DisplayName != "" {
		return u.DisplayName
	}
	if u.GivenName != "" && u.Surname != "" {
		return u.GivenName + " " + u.Surname
	}
	return ""
}
