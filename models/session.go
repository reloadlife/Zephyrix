package models

import "time"

type SessionEntity struct {
	ID        uint64 `orm:"table=zephyrix_sessions"`
	SessionID string `orm:"unique=session_id"`

	UserID uint64 `orm:"index=user_id"`
	Data   string `orm:"type:text"`

	CreatedAt time.Time `orm:"time"`
	ExpiresAt time.Time `orm:"time"`
}
