package zephyrix

import "errors"

var (
	ErrRateLimited      = errors.New("rate limit exceeded")
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrMFARequired      = errors.New("MFA required")
	ErrInvalidMFAToken  = errors.New("invalid MFA token")
	ErrUnsupportedOAuth = errors.New("unsupported OAuth2 operation")
)
