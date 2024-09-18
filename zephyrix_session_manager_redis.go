package zephyrix

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/latolukasz/beeorm/v3"
)

type RedisSessionStorage struct {
	client beeorm.RedisCache
	orm    beeorm.ORM
	prefix string
}

func NewRedisSessionStorage(orm beeorm.ORM, client beeorm.RedisCache, prefix string) *RedisSessionStorage {
	return &RedisSessionStorage{
		client: client,
		prefix: prefix,
		orm:    orm,
	}
}

func (r *RedisSessionStorage) Create(ctx context.Context, session *Session) error {
	return r.Update(ctx, session)
}

func (r *RedisSessionStorage) Get(ctx context.Context, sessionID string) (*Session, error) {
	key := r.getKey(sessionID)
	data, has := r.client.Get(r.orm, key)
	if !has {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	var session Session
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *RedisSessionStorage) Update(ctx context.Context, session *Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	key := r.getKey(session.ID)
	expiration := session.ExpiresAt.Sub(time.Now())
	r.client.Set(r.orm, key, data, expiration)
	return nil
}

func (r *RedisSessionStorage) Delete(ctx context.Context, sessionID string) error {
	key := r.getKey(sessionID)
	r.client.Del(r.orm, key)
	return nil
}

func (r *RedisSessionStorage) Cleanup(ctx context.Context) error {
	// Redis automatically removes expired keys, so no cleanup is needed
	return nil
}

func (r *RedisSessionStorage) getKey(sessionID string) string {
	return fmt.Sprintf("%s:session:%s", r.prefix, sessionID)
}
