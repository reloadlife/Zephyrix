package zephyrix

import (
	"context"
	"fmt"
	"time"
)

type User interface {
	// Basic user information
	ID() uint64
	Username() string
	Email() string

	// Password management
	CheckPassword(password string) bool
	SetPassword(password string) error
	PasswordLastChanged() time.Time

	// Role management
	Roles() []string
	HasRole(role string) bool
	AddRole(role string) error
	RemoveRole(role string) error

	// MFA management
	HasMFAEnabled() bool
	EnabledMFAMethods() []MFAMethod
	SetupMFA(method MFAMethod, secret string) error
	DisableMFA(method MFAMethod) error
	GetMFASecret(method MFAMethod) (string, error)

	// TOTP specific methods
	SetTOTPSecret(secret string) error
	GetTOTPSecret() (string, error)

	// Account status
	IsActive() bool
	IsLocked() bool
	Lock() error
	Unlock() error

	// Account metadata
	CreatedAt() time.Time
	UpdatedAt() time.Time
	LastLoginAt() time.Time
	SetLastLoginAt(time time.Time) error

	// Custom data storage (for flexibility)
	SetMetadata(key string, value interface{}) error
	GetMetadata(key string) (interface{}, error)
}

type GrantType string

const (
	GrantTypePassword      GrantType = "password"
	GrantTypeRefreshToken  GrantType = "refresh_token"
	GrantTypeAuthorization GrantType = "authorization_code"
	GrantTypeMagicToken    GrantType = "magic_token" // AKA "magic link"
)

type Authenticate struct {
	GrantType GrantType

	Auth     string
	Password string

	RefreshToken string

	Provider string
	Token    string
	Code     string

	MagicToken string
}

func (ap *AuthProvider) AuthenticateUser(ctx context.Context, input Authenticate) (string, error) {
	// Rate limiting
	rl := ap.components.rateLimiter.Limiter(ctx, "login")
	rlKey := fmt.Sprintf("login:%s:%s:%s:%s", input.Auth, input.RefreshToken, input.Provider+input.Token, input.MagicToken)
	if !rl.Allow(ctx, "login", rlKey) {
		return "", ErrRateLimited
	}

	var user User
	var err error

	switch input.GrantType {
	case GrantTypePassword:
		user, err = ap.authenticateWithPassword(ctx, input.Auth, input.Password)
	case GrantTypeRefreshToken:
		user, err = ap.authenticateWithRefreshToken(ctx, input.RefreshToken)
	case GrantTypeAuthorization:
		user, err = ap.authenticateWithAuthorizationCode(ctx, input.Provider, input.Code)
	case GrantTypeMagicToken:
		user, err = ap.authenticateWithMagicToken(ctx, input.MagicToken)
	default:
		return "", fmt.Errorf("unsupported grant type: %s", input.GrantType)
	}

	if err != nil {
		return "", fmt.Errorf("authentication failed: %w", err)
	}

	// Check if MFA is required
	if ap.config.MFA.Enabled && ap.components.mfaManager.IsRequired(user) {
		return "", ErrMFARequired
	}

	// Generate JWT token
	token, err := ap.generateJWT(user)
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT: %w", err)
	}

	// Log successful login
	ap.components.auditLogger.logSuccessfulLogin(ctx, user.Username())
	return token, nil
}

// Helper functions for each authentication method
func (ap *AuthProvider) authenticateWithPassword(ctx context.Context, auth, password string) (User, error) {
	// Implement password-based authentication
	// This might involve querying a database, checking password hashes, etc.
	// Return the authenticated user or an error if authentication fails
	return nil, fmt.Errorf("not implemented")
}

func (ap *AuthProvider) authenticateWithRefreshToken(ctx context.Context, refreshToken string) (User, error) {
	// Implement refresh token-based authentication
	// This might involve validating the refresh token, fetching the associated user, etc.
	// Return the authenticated user or an error if authentication fails
	return nil, fmt.Errorf("not implemented")
}

func (ap *AuthProvider) authenticateWithAuthorizationCode(ctx context.Context, provider, code string) (User, error) {
	// Implement authorization code-based authentication (e.g., OAuth)
	// This might involve exchanging the code for tokens, validating the provider, fetching user info, etc.
	// Return the authenticated user or an error if authentication fails
	return nil, fmt.Errorf("not implemented")
}

func (ap *AuthProvider) authenticateWithMagicToken(ctx context.Context, magicToken string) (User, error) {
	// Implement magic token (magic link) based authentication
	// This might involve validating the token, fetching the associated user, etc.
	// Return the authenticated user or an error if authentication fails

	return nil, fmt.Errorf("not implemented")
}
