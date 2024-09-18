package zephyrix

import (
	"time"
)

type AuthConfig struct {
	RedisPool           string             `mapstructure:"redis_pool"`
	GrantTypes          []string           `mapstructure:"grant_types"`
	APIKey              APIKeyConfig       `mapstructure:"api_key"`
	JWT                 JWTConfig          `mapstructure:"jwt"`
	Session             SessionConfig      `mapstructure:"session"`
	OAuth2              OAuth2Config       `mapstructure:"oauth2"`
	MFA                 MFAConfig          `mapstructure:"multi_factor_authentication"`
	PasswordPolicy      PasswordPolicy     `mapstructure:"password_policy"`
	AccountLockout      AccountLockout     `mapstructure:"account_lockout"`
	RateLimiting        RateLimitingConfig `mapstructure:"rate_limiting"`
	SecurityHeaders     SecurityHeaders    `mapstructure:"security_headers"`
	UserRegistration    UserRegistration   `mapstructure:"user_registration"`
	AccountRecovery     AccountRecovery    `mapstructure:"account_recovery"`
	SessionManagement   SessionManagement  `mapstructure:"session_management"`
	ConsentManagement   ConsentManagement  `mapstructure:"consent_management"`
	Passwordless        Passwordless       `mapstructure:"passwordless"`
	Geofencing          Geofencing         `mapstructure:"geofencing"`
	Webhooks            WebhooksConfig     `mapstructure:"webhooks"`
	FeatureToggles      FeatureToggles     `mapstructure:"feature_toggles"`
	PasswordHashingCost int                `mapstructure:"hashing_cost"`
	PasswordHashingSalt string             `mapstructure:"hashing_salt"`
	Audit               bool               `mapstructure:"audit"`
}

type APIKeyConfig struct {
	Enabled    bool          `mapstructure:"enabled"`
	Expiration time.Duration `mapstructure:"expiration"`
}

type JWTConfig struct {
	Secret            string        `mapstructure:"secret"`
	Issuer            string        `mapstructure:"issuer"`
	Audience          string        `mapstructure:"audience"`
	Expiration        time.Duration `mapstructure:"expiration"`
	RefreshExpiration time.Duration `mapstructure:"refresh_expiration"`
	SigningMethod     string        `mapstructure:"signing_method"`
}


type OAuth2Provider struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url"`
}

type PasswordPolicy struct {
	MinLength           int           `mapstructure:"min_length"`
	RequireUppercase    bool          `mapstructure:"require_uppercase"`
	RequireLowercase    bool          `mapstructure:"require_lowercase"`
	RequireNumbers      bool          `mapstructure:"require_numbers"`
	RequireSpecialChars bool          `mapstructure:"require_special_chars"`
	MaxAge              time.Duration `mapstructure:"max_age"`
	HistoryCount        int           `mapstructure:"history_count"`
}

type AccountLockout struct {
	MaxAttempts     int           `mapstructure:"max_attempts"`
	LockoutDuration time.Duration `mapstructure:"lockout_duration"`
	ResetAfter      time.Duration `mapstructure:"reset_after"`
}

type RateLimitingConfig struct {
	Enabled       bool              `mapstructure:"enabled"`
	LoginAttempts RateLimitSettings `mapstructure:"login_attempts"`
	PasswordReset RateLimitSettings `mapstructure:"password_reset"`
}

type RateLimitSettings struct {
	Max int           `mapstructure:"max"`
	Per time.Duration `mapstructure:"per"`
}

type SecurityHeaders struct {
	HSTSEnabled   bool   `mapstructure:"hsts_enabled"`
	CSPEnabled    bool   `mapstructure:"csp_enabled"`
	XFrameOptions string `mapstructure:"xframe_options"`
}

type UserRegistration struct {
	EmailVerificationRequired bool     `mapstructure:"email_verification_required"`
	InvitationOnly            bool     `mapstructure:"invitation_only"`
	AllowedDomains            []string `mapstructure:"allowed_domains"`
}

type AccountRecovery struct {
	Methods         []string      `mapstructure:"methods"`
	TokenExpiration time.Duration `mapstructure:"token_expiration"`
}

type SessionManagement struct {
	ForceLogoutAll bool `mapstructure:"force_logout_all"`
}

type ConsentManagement struct {
	RequireConsent  bool          `mapstructure:"require_consent"`
	ConsentValidity time.Duration `mapstructure:"consent_validity"`
}

type Passwordless struct {
	Enabled bool     `mapstructure:"enabled"`
	Methods []string `mapstructure:"methods"`
}

type Geofencing struct {
	Enabled          bool     `mapstructure:"enabled"`
	AllowedCountries []string `mapstructure:"allowed_countries"`
}

type WebhooksConfig struct {
	LoginSuccess string `mapstructure:"login_success"`
	LoginFailure string `mapstructure:"login_failure"`
}

type FeatureToggles struct {
	SocialLogin  bool `mapstructure:"social_login"`
	APIKeyAuth   bool `mapstructure:"api_key_auth"`
	MFA          bool `mapstructure:"mfa"`
	Passwordless bool `mapstructure:"passwordless"`
}
