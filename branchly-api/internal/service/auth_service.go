package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/branchly/branchly-api/internal/config"
	"github.com/branchly/branchly-api/internal/domain"
	"github.com/branchly/branchly-api/internal/infra"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type AuthService struct {
	cfg   *config.Config
	users domain.UserRepository
	oauth *oauth2.Config
}

type jwtClaims struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	jwt.RegisteredClaims
}

func NewAuthService(cfg *config.Config, users domain.UserRepository) *AuthService {
	o := &oauth2.Config{
		ClientID:     cfg.GitHubClientID,
		ClientSecret: cfg.GitHubClientSecret,
		RedirectURL:  cfg.GitHubRedirectURI,
		Scopes:   []string{"repo", "user:email"},
		Endpoint: github.Endpoint,
	}
	return &AuthService{cfg: cfg, users: users, oauth: o}
}

func (s *AuthService) OAuthAuthCodeURL(state string) string {
	return s.oauth.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (s *AuthService) ExchangeGitHubCode(ctx context.Context, code string) (*oauth2.Token, error) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	tok, err := s.oauth.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("auth service: exchange github code: %w", err)
	}
	return tok, nil
}

type githubUserJSON struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

func (s *AuthService) FetchGitHubProfile(ctx context.Context, accessToken string) (*githubUserJSON, string, error) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	client := &http.Client{Timeout: 8 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return nil, "", fmt.Errorf("auth service: github user request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	res, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("auth service: github user http: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("auth service: github user status %d", res.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return nil, "", fmt.Errorf("auth service: read github user: %w", err)
	}
	var u githubUserJSON
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, "", fmt.Errorf("auth service: decode github user: %w", err)
	}
	email := u.Email
	if email == "" {
		em, err := s.fetchPrimaryGitHubEmail(ctx, accessToken)
		if err != nil {
			return nil, "", fmt.Errorf("auth service: github emails: %w", err)
		}
		email = em
	}
	if email == "" {
		email = fmt.Sprintf("%s@users.noreply.github.com", u.Login)
	}
	return &u, email, nil
}

type githubEmailEntry struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func (s *AuthService) fetchPrimaryGitHubEmail(ctx context.Context, accessToken string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	cli := &http.Client{Timeout: 8 * time.Second}
	res, err := cli.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", res.StatusCode)
	}
	var entries []githubEmailEntry
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&entries); err != nil {
		return "", err
	}
	for _, e := range entries {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	for _, e := range entries {
		if e.Verified {
			return e.Email, nil
		}
	}
	return "", nil
}

func (s *AuthService) UpsertGitHubUser(ctx context.Context, accessToken string, gh *githubUserJSON, email string) (*domain.User, error) {
	enc, err := infra.Encrypt(accessToken, s.cfg.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("auth service: encrypt token: %w", err)
	}
	name := gh.Name
	if name == "" {
		name = gh.Login
	}
	u := &domain.User{
		Provider:       "github",
		ProviderID:     fmt.Sprintf("%d", gh.ID),
		Email:          email,
		Name:           name,
		AvatarURL:      gh.AvatarURL,
		EncryptedToken: enc,
	}
	out, err := s.users.UpsertByProvider(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("auth service: upsert user: %w", err)
	}
	return out, nil
}

func (s *AuthService) IssueAccessToken(user *domain.User) (string, error) {
	now := time.Now()
	claims := jwtClaims{
		Email: user.Email,
		Name:  user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.JWTTTL())),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := t.SignedString(s.cfg.JWTSecret)
	if err != nil {
		return "", fmt.Errorf("auth service: sign jwt: %w", err)
	}
	return signed, nil
}

func (s *AuthService) ValidateAccessToken(tokenString string) (subject string, err error) {
	tok, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.cfg.JWTSecret, nil
	})
	if err != nil {
		return "", fmt.Errorf("auth service: parse jwt: %w", err)
	}
	claims, ok := tok.Claims.(*jwtClaims)
	if !ok || !tok.Valid {
		return "", fmt.Errorf("auth service: invalid claims")
	}
	if claims.Subject == "" {
		return "", fmt.Errorf("auth service: empty subject")
	}
	return claims.Subject, nil
}

func (s *AuthService) Me(ctx context.Context, userID string) (*domain.User, error) {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("auth service: me: %w", err)
	}
	if u == nil {
		return nil, nil
	}
	return u, nil
}
