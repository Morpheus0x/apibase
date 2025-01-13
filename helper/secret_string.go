package helper

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type SecretString struct {
	value string
}

func CreateSecretString(value string) SecretString {
	return SecretString{value: value}
}

func (s SecretString) String() string {
	return "*****"
}

func (s SecretString) GetSecret() string {
	return s.value
}

// Used by BurntSushi/toml to parse secrets from .toml config file
func (s *SecretString) UnmarshalText(text []byte) error {
	s.value = string(text)
	return nil
}

// Used to get value from json parser
func (s *SecretString) UnmarshalJSON(b []byte) error {
	// In JSON, the secret is a string value
	var tmp string
	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}
	s.value = tmp
	return nil
}

// Used to parse secret string to json
func (s SecretString) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.value)
}

// Used by pgx to scan value from database
func (s *SecretString) Scan(src interface{}) error {
	if src == nil {
		s.value = ""
		return nil
	}
	switch v := src.(type) {
	case string:
		s.value = v
	case []byte:
		s.value = string(v)
	default:
		return fmt.Errorf("cannot convert %T to SecretString", src)
	}
	return nil
}

// Used by pgx to get value for db update/insert/select...where
func (s SecretString) Value() (driver.Value, error) {
	return s.value, nil
}
