package models

import "time"

type AuditLogEntity struct {
	ID uint64 `orm:"table=audit_logs;mysql=audit"`

	UserID string `orm:"index=user_id"`
	Action string `orm:"index=action"`

	Details string `orm:"length=max"`

	CreatedAt time.Time `orm:"time"`
}
