package zephyrix

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/latolukasz/beeorm/v3"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

type OAuth2Config struct {
	ProvidersSource string              `mapstructure:"providers_source"`
	Providers       map[string]Provider `mapstructure:"providers"`
}

type Provider struct {
	ClientID     string   `mapstructure:"client_id"`
	ClientSecret string   `mapstructure:"client_secret"`
	RedirectURL  string   `mapstructure:"redirect_url"`
	Scopes       []string `mapstructure:"scopes"`
}

type OAuth2Manager struct {
	config    OAuth2Config
	orm       beeorm.Engine
	providers map[string]*oauth2.Config
	mu        sync.RWMutex
}

func NewOAuth2Manager(config OAuth2Config, orm beeorm.Engine) *OAuth2Manager {
	manager := &OAuth2Manager{
		config:    config,
		orm:       orm,
		providers: make(map[string]*oauth2.Config),
	}

	return manager
}

func (om *OAuth2Manager) Initialize(ctx context.Context) error {
	if om.config.ProvidersSource == "database" {
		return om.loadProvidersFromDatabase(ctx)
	}
	return om.loadProvidersFromConfig()
}

func (om *OAuth2Manager) loadProvidersFromConfig() error {
	for name, provider := range om.config.Providers {
		if err := om.AddProvider(name, provider); err != nil {
			return fmt.Errorf("failed to add provider %s: %w", name, err)
		}
	}
	return nil
}

func (om *OAuth2Manager) loadProvidersFromDatabase(ctx context.Context) error {
	// Implement loading providers from the database
	// This is a placeholder implementation
	return errors.New("loading providers from database not implemented")
}

func (om *OAuth2Manager) AddProvider(name string, provider Provider) error {
	om.mu.Lock()
	defer om.mu.Unlock()

	var endpoint oauth2.Endpoint
	switch name {
	case "github":
		endpoint = github.Endpoint
	case "google":
		endpoint = google.Endpoint
	default:
		return fmt.Errorf("unsupported provider: %s", name)
	}

	om.providers[name] = &oauth2.Config{
		ClientID:     provider.ClientID,
		ClientSecret: provider.ClientSecret,
		RedirectURL:  provider.RedirectURL,
		Scopes:       provider.Scopes,
		Endpoint:     endpoint,
	}

	return nil
}

func (om *OAuth2Manager) RemoveProvider(name string) {
	om.mu.Lock()
	defer om.mu.Unlock()
	delete(om.providers, name)
}

func (om *OAuth2Manager) GetAuthURL(providerName string, state string) (string, error) {
	om.mu.RLock()
	defer om.mu.RUnlock()

	provider, ok := om.providers[providerName]
	if !ok {
		return "", fmt.Errorf("provider not found: %s", providerName)
	}

	return provider.AuthCodeURL(state), nil
}

func (om *OAuth2Manager) Exchange(ctx context.Context, providerName string, code string) (*oauth2.Token, error) {
	om.mu.RLock()
	defer om.mu.RUnlock()

	provider, ok := om.providers[providerName]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", providerName)
	}

	return provider.Exchange(ctx, code)
}

func (om *OAuth2Manager) GetUserInfo(ctx context.Context, providerName string, token *oauth2.Token) (map[string]interface{}, error) {
	om.mu.RLock()
	provider, ok := om.providers[providerName]
	om.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("provider not found: %s", providerName)
	}

	client := provider.Client(ctx, token)
	resp, err := client.Get(om.getUserInfoURL(providerName))
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: status code %d", resp.StatusCode)
	}

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return userInfo, nil
}

func (om *OAuth2Manager) RefreshToken(ctx context.Context, providerName string, token *oauth2.Token) (*oauth2.Token, error) {
	om.mu.RLock()
	provider, ok := om.providers[providerName]
	om.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("provider not found: %s", providerName)
	}

	if !token.Expiry.IsZero() && token.Expiry.After(time.Now()) {
		return token, nil
	}

	src := provider.TokenSource(ctx, token)
	newToken, err := src.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	return newToken, nil
}

func (om *OAuth2Manager) getUserInfoURL(providerName string) string {
	switch providerName {
	case "github":
		return "https://api.github.com/user"
	case "google":
		return "https://www.googleapis.com/oauth2/v2/userinfo"
	default:
		return ""
	}
}

func (om *OAuth2Manager) SyncProviders(ctx context.Context) error {
	if om.config.ProvidersSource != "database" {
		return nil
	}

	// Implement syncing providers with the database
	// This is a placeholder implementation
	return errors.New("syncing providers with database not implemented")
}
