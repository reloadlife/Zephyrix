package zephyrix

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/latolukasz/beeorm/v3"
	"go.uber.org/fx"
)

type SessionConfig struct {
	StorageType     string        `mapstructure:"storage_type"` // "redis", "mysql", or "memory"
	Pool            string        `mapstructure:"pool"`
	Prefix          string        `mapstructure:"prefix"`
	Expiration      time.Duration `mapstructure:"expiration"`
	RefreshWindow   time.Duration `mapstructure:"refresh_window"`
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"`
}

type Session struct {
	ID        string                 `json:"id"`
	UserID    uint64                 `json:"user_id"`
	CreatedAt time.Time              `json:"created_at"`
	ExpiresAt time.Time              `json:"expires_at"`
	Data      map[string]interface{} `json:"data"`
}

type SessionStorage interface {
	Create(ctx context.Context, session *Session) error
	Get(ctx context.Context, sessionID string) (*Session, error)
	Update(ctx context.Context, session *Session) error
	Delete(ctx context.Context, sessionID string) error
	Cleanup(ctx context.Context) error
}

type SessionManager struct {
	config  SessionConfig
	storage SessionStorage
}

func NewSessionManager(lc fx.Lifecycle, conf *Config, orm beeorm.Engine) (*SessionManager, error) {
	var storage SessionStorage
	config := conf.Authentication.Session

	switch config.StorageType {
	case "redis":
	default:
		storage = NewRedisSessionStorage(orm.NewORM(context.Background()), orm.Redis(config.Pool), config.Prefix)
	case "mysql":
		storage = NewMySQLSessionStorage(orm, config.Prefix)
	case "memory":
		storage = NewMemorySessionStorage()
		return nil, fmt.Errorf("unsupported storage type: %s", config.StorageType)
	}

	sm := &SessionManager{
		config:  config,
		storage: storage,
	}

	// Start the cleanup task
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			sm.StartCleanupTask(ctx)
			return nil
		},
	})

	return sm, nil
}

func (sm *SessionManager) CreateSession(ctx context.Context, userID uint64) (*Session, error) {
	sessionID := uuid.New().String()
	now := time.Now()
	session := &Session{
		ID:        sessionID,
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: now.Add(sm.config.Expiration),
		Data:      make(map[string]interface{}),
	}

	if err := sm.storage.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

func (sm *SessionManager) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	session, err := sm.storage.Get(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if time.Now().After(session.ExpiresAt) {
		sm.DestroySession(ctx, sessionID)
		return nil, fmt.Errorf("session expired")
	}

	if sm.shouldRefreshSession(*session) {
		if err := sm.refreshSession(ctx, session); err != nil {
			return nil, fmt.Errorf("failed to refresh session: %w", err)
		}
	}

	return session, nil
}

func (sm *SessionManager) shouldRefreshSession(session Session) bool {
	return time.Now().Add(sm.config.RefreshWindow).After(session.ExpiresAt)
}

func (sm *SessionManager) refreshSession(ctx context.Context, session *Session) error {
	session.ExpiresAt = time.Now().Add(sm.config.Expiration)
	return sm.storage.Update(ctx, session)
}

func (sm *SessionManager) UpdateSessionData(ctx context.Context, sessionID string, data map[string]interface{}) error {
	session, err := sm.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	for k, v := range data {
		session.Data[k] = v
	}

	return sm.storage.Update(ctx, session)
}

func (sm *SessionManager) DestroySession(ctx context.Context, sessionID string) error {
	return sm.storage.Delete(ctx, sessionID)
}

func (sm *SessionManager) StartCleanupTask(ctx context.Context) {
	ticker := time.NewTicker(sm.config.CleanupInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := sm.storage.Cleanup(ctx); err != nil {
					Logger.Error("failed to cleanup sessions: %v", err)
				}
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}
