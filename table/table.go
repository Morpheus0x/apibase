package table

import (
	"time"

	h "gopkg.cc/apibase/helper"
)

type User struct {
	ID             int            `db:"id" default:"true" table:"users"`
	Name           string         `db:"name"`
	AuthProvider   string         `db:"auth_provider"`
	Email          string         `db:"email"`
	EmailVerified  bool           `db:"email_verified"`
	PasswordHash   h.SecretString `db:"password_hash"`
	SecretsVersion int            `db:"secrets_version"`
	TotpSecret     string         `db:"totp_secret"`
	SuperAdmin     bool           `db:"super_admin"`
	CreatedAt      time.Time      `db:"created_at" default:"true"`
	UpdatedAt      time.Time      `db:"updated_at" default:"true"`
}

type RefreshToken struct {
	ID           int            `db:"id" default:"true" table:"refresh_tokens"`
	UserID       int            `db:"user_id"`
	SessionID    h.SecretString `db:"session_id"`
	ReissueCount int            `db:"reissue_count"`
	UserAgent    string         `db:"user_agent"`
	ExpiresAt    time.Time      `db:"expires_at"`
	CreatedAt    time.Time      `db:"created_at" default:"true"`
	UpdatedAt    time.Time      `db:"updated_at" default:"true"`
}

type Organization struct {
	ID          int    `db:"id" default:"true" table:"organizations"`
	Name        string `db:"name"`
	Description string `db:"description"`
}

type UserRole struct {
	ID       int  `db:"id" default:"true" table:"user_roles"`
	UserID   int  `db:"user_id"`
	OrgID    int  `db:"org_id"`
	OrgView  bool `db:"org_view"`
	OrgEdit  bool `db:"org_edit"`
	OrgAdmin bool `db:"org_admin"`
}

type ScheduledTask struct {
	ID        int       `db:"id" default:"true" table:"scheduled_tasks"`
	TaskID    string    `db:"task_id"`
	OrgID     int       `db:"org_id"`
	StartDate time.Time `db:"start_date"`
	Interval  Duration  `db:"interval"`
	TaskType  string    `db:"task_type"`
	TaskData  string    `db:"task_data"`
	CreatedAt time.Time `db:"created_at" default:"true"`
	UpdatedAt time.Time `db:"updated_at" default:"true"`
}
