package zephyrix

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/latolukasz/beeorm/v3"
	"go.uber.org/fx"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
)

type AuthProvider struct {
	config      *AuthConfig
	orm         beeorm.Engine
	redisClient beeorm.RedisCache
	jwtSecret   []byte
	components  struct {
		passwordHasher int
		passwordSalt   string
		auditLogger    *AuditLogger
		rateLimiter    *RateLimiter
		mfaManager     *MFAManager
		oauth2Manager  *OAuth2Manager
		sessionManager *SessionManager
	}
	providerCache sync.Map
}

func NewAuthProvider(lc fx.Lifecycle, conf *Config, orm beeorm.Engine, redisClient beeorm.RedisCache, a *AuditLogger, rl *RateLimiter) (*AuthProvider, error) {
	ap := &AuthProvider{
		config:      &conf.Authentication,
		orm:         orm,
		redisClient: redisClient,
		jwtSecret:   []byte(conf.Authentication.JWT.Secret),
	}

	ap.components.passwordHasher = bcrypt.DefaultCost
	ap.components.passwordSalt = conf.Authentication.PasswordHashingSalt
	ap.components.auditLogger = a
	ap.components.rateLimiter = rl

	lc.Append(fx.Hook{
		OnStart: ap.initialize,
		OnStop:  ap.cleanup,
	})

	return ap, nil
}

func (ap *AuthProvider) initialize(ctx context.Context) error {
	if ap.config.OAuth2.ProvidersSource == "database" {
		if err := ap.components.oauth2Manager.SyncProviders(ctx); err != nil {
			return fmt.Errorf("failed to sync OAuth2 providers: %w", err)
		}
	}

	return ap.warmCaches(ctx)
}

func (ap *AuthProvider) cleanup(ctx context.Context) error {
	return ap.cleanupTemporaryData(ctx)
}

func (ap *AuthProvider) VerifyMFA(ctx context.Context, username, method, token string) (string, error) {
	user, err := ap.getUserByUsername(ctx, username)
	if err != nil {
		ap.components.auditLogger.logFailedLogin(ctx, username, "user_not_found")
		return "", ErrUserNotFound
	}

	valid, err := ap.components.mfaManager.VerifyMFA(ctx, user, MFAMethod(method), token)
	if err != nil {
		return "", fmt.Errorf("MFA verification failed: %w", err)
	}

	if !valid {
		ap.components.auditLogger.logFailedMFA(ctx, username, method)
		return "", ErrInvalidMFAToken
	}

	jwtToken, err := ap.generateJWT(user)
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT after MFA: %w", err)
	}

	ap.components.auditLogger.logSuccessfulMFA(ctx, username, method)
	return jwtToken, nil
}

func (ap *AuthProvider) InitializeOAuth2(ctx context.Context) error {
	return ap.components.oauth2Manager.Initialize(ctx)
}

func (ap *AuthProvider) GetOAuth2AuthURL(providerName, state string) (string, error) {
	return ap.components.oauth2Manager.GetAuthURL(providerName, state)
}

func (ap *AuthProvider) HandleOAuth2Callback(ctx context.Context, providerName, code string) (User, error) {
	token, err := ap.components.oauth2Manager.Exchange(ctx, providerName, code)
	if err != nil {
		return nil, fmt.Errorf("OAuth2 code exchange failed: %w", err)
	}

	userInfo, err := ap.components.oauth2Manager.GetUserInfo(ctx, providerName, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth2 user info: %w", err)
	}

	user, err := ap.findOrCreateOAuth2User(ctx, providerName, userInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to find or create OAuth2 user: %w", err)
	}

	if err := ap.storeOAuth2Token(ctx, user, providerName, token); err != nil {
		return nil, fmt.Errorf("failed to store OAuth2 token: %w", err)
	}

	return user, nil
}

func (ap *AuthProvider) generateJWT(user User, opts ...jwt.MapClaims) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": user.ID(),
		"exp": now.Add(ap.config.JWT.Expiration).Unix(),
		"iat": now.Unix(),
		"iss": ap.config.JWT.Issuer,
		"aud": ap.config.JWT.Audience,
	}

	for _, opt := range opts {
		for k, v := range opt {
			claims[k] = v
		}
	}

	token := jwt.NewWithClaims(ap.getSigningMethod(), claims)
	return token.SignedString(ap.jwtSecret)
}

func (ap *AuthProvider) VerifyToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return ap.jwtSecret, nil
	})
}

func (ap *AuthProvider) getSigningMethod() jwt.SigningMethod {
	switch ap.config.JWT.SigningMethod {
	case "HS384":
		return jwt.SigningMethodHS384
	case "HS512":
		return jwt.SigningMethodHS512
	default:
		return jwt.SigningMethodHS256
	}
}

func (ap *AuthProvider) getUserByUsername(ctx context.Context, username string) (User, error) {
	// Implementation depends on your User model and ORM usage
	fmt.Println("getUserByUsername not implemented")

	return nil, nil
}

func (ap *AuthProvider) warmCaches(ctx context.Context) error {
	// Implementation for warming up caches
	fmt.Println("warmCaches not implemented")

	return nil
}

func (ap *AuthProvider) cleanupTemporaryData(ctx context.Context) error {
	// Implementation for cleaning up temporary data
	fmt.Println("cleanupTemporaryData not implemented")

	return nil
}

func (ap *AuthProvider) findOrCreateOAuth2User(ctx context.Context, providerName string, userInfo map[string]interface{}) (User, error) {
	// Implementation for finding or creating a user based on OAuth2 user info
	fmt.Println("findOrCreateOAuth2User not implemented")

	return nil, nil
}

func (ap *AuthProvider) storeOAuth2Token(ctx context.Context, user User, providerName string, token *oauth2.Token) error {
	// Implementation for storing the OAuth2 token for the user
	fmt.Println("storeOAuth2Token not implemented")

	return nil
}

// AuthProviderModule provides an fx.Option to register the AuthProvider with the application
func AuthProviderModule() fx.Option {
	return fx.Module(
		"auth_provider",
		fx.Provide(NewAuthProvider),
		fx.Invoke(func(*AuthProvider) {}),
	)
}
