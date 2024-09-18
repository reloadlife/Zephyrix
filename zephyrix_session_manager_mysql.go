package zephyrix

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/latolukasz/beeorm/v3"
	"go.mamad.dev/zephyrix/models"
)

type MySQLSessionStorage struct {
	orm    beeorm.Engine
	prefix string
}

func NewMySQLSessionStorage(orm beeorm.Engine, prefix string) *MySQLSessionStorage {
	return &MySQLSessionStorage{
		orm:    orm,
		prefix: prefix,
	}
}

func (m *MySQLSessionStorage) Create(ctx context.Context, session *Session) error {
	orm := m.orm.NewORM(ctx)
	sess := beeorm.NewEntity[models.SessionEntity](orm)
	sess.SessionID = session.ID
	sess.UserID = session.UserID
	sess.ExpiresAt = session.ExpiresAt.UTC()
	data, _ := json.Marshal(session.Data)
	sess.Data = string(data)
	return orm.FlushAsync()
}

func (m *MySQLSessionStorage) Get(ctx context.Context, sessionID string) (*Session, error) {
	orm := m.orm.NewORM(ctx)
	sess, found := beeorm.GetByUniqueIndex[models.SessionEntity](orm, "session_id", sessionID)
	if !found {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	data := map[string]interface{}{}
	_ = json.Unmarshal([]byte(sess.Data), &data)

	return &Session{
		ID:        sess.SessionID,
		UserID:    sess.UserID,
		CreatedAt: sess.CreatedAt,
		ExpiresAt: sess.ExpiresAt,
		Data:      data,
	}, nil
}

func (m *MySQLSessionStorage) Update(ctx context.Context, session *Session) error {
	orm := m.orm.NewORM(ctx)

	sess, found := beeorm.GetByUniqueIndex[models.SessionEntity](orm, "session_id", session.ID)
	if !found {
		return fmt.Errorf("session not found: %s", session.ID)
	}

	sess = beeorm.EditEntity(orm, sess)

	sess.UserID = session.UserID
	sess.ExpiresAt = session.ExpiresAt.UTC()
	data, _ := json.Marshal(session.Data)
	sess.Data = string(data)

	return orm.FlushAsync()
}

func (m *MySQLSessionStorage) Delete(ctx context.Context, sessionID string) error {
	orm := m.orm.NewORM(ctx)

	sess, found := beeorm.GetByUniqueIndex[models.SessionEntity](orm, "session_id", sessionID)
	if !found {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	beeorm.DeleteEntity(orm, sess)

	return orm.FlushAsync()
}

func (m *MySQLSessionStorage) Cleanup(ctx context.Context) error {
	orm := m.orm.NewORM(ctx)

	iterator := beeorm.Search[models.SessionEntity](orm, beeorm.NewWhere("ExpiresAt", "<", time.Now().UTC()), nil)
	for iterator.Next() {
		beeorm.DeleteEntity(orm, iterator.Entity())
	}

	return orm.FlushAsync()
}
