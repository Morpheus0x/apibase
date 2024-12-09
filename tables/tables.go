package tables

import "time"

type Users struct {
	ID             int       `db:"id" default:"true" table:"users"`
	Name           string    `db:"name"`
	Role           string    `db:"role"`
	AuthProvider   string    `db:"auth_provider"`
	Email          string    `db:"email"`
	EmailVerified  bool      `db:"email_verified"`
	PasswordHash   string    `db:"password_hash"`
	SecretsVersion int       `db:"secrets_version"`
	TotpSecret     string    `db:"totp_secret"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

type RefreshTokens struct {
	ID           int       `db:"id" default:"true" table:"refresh_tokens"`
	UserID       int       `db:"user_id"`
	Token        string    `db:"token"`
	ReissueCount int       `db:"reissue_count"`
	CreatedAt    time.Time `db:"created_at" default:"true"`
	UpdatedAt    time.Time `db:"updated_at" default:"true"`
	ExpiresAt    time.Time `db:"expires_at"`
}
