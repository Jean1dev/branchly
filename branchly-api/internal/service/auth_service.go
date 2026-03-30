package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/branchly/branchly-api/internal/config"
	"github.com/branchly/branchly-api/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	cfg      *config.Config
	users    domain.UserRepository
	integSvc *IntegrationService
}

type jwtClaims struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	jwt.RegisteredClaims
}

func NewAuthService(cfg *config.Config, users domain.UserRepository, integSvc *IntegrationService) *AuthService {
	return &AuthService{cfg: cfg, users: users, integSvc: integSvc}
}

func (s *AuthService) UpsertFromInternalOAuth(
	ctx context.Context,
	providerID, email, name, avatarURL, githubToken string,
) (*domain.User, error) {
	displayName := strings.TrimSpace(name)
	if displayName == "" {
		displayName = "GitHub user"
	}
	em := strings.TrimSpace(email)
	if em == "" {
		em = fmt.Sprintf("github-%s@users.noreply.github.com", providerID)
	}
	// User document no longer stores the encrypted token — token lives in git_integrations.
	u := &domain.User{
		Provider:  "github",
		ProviderID: providerID,
		Email:     em,
		Name:      displayName,
		AvatarURL: strings.TrimSpace(avatarURL),
	}
	out, err := s.users.UpsertByProvider(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("auth service: upsert user: %w", err)
	}

	// Always refresh the GitHub integration token on sign-in.
	if _, err := s.integSvc.ConnectGitHub(ctx, out.ID, githubToken); err != nil {
		return nil, fmt.Errorf("auth service: connect github: %w", err)
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
