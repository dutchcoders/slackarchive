package utils

import (
	"database/sql/driver"

	"github.com/nu7hatch/gouuid"
)

type UUID []byte

func NewNullUUID() NullUUID {
	u4, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}

	return NullUUID{Parse(u4.String()), true}
}

func NewUUID() UUID {
	u4, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}

	return Parse(u4.String())
}

type NullUUID struct {
	UUID  UUID
	Valid bool
}

func (u *NullUUID) Scan(value interface{}) error {
	if value == nil {
		u.UUID, u.Valid = nil, false
		return nil
	}

	u.Valid = true

	b := value.([]byte)
	u.UUID = make([]byte, len(b))
	copy(u.UUID, b)
	return nil
}

func (u NullUUID) Value() (driver.Value, error) {
	if !u.Valid {
		return nil, nil
	}
	return u.UUID.Value()
}

func Parse(s string) UUID {
	return UUID(s)
}

func (u UUID) MarshalText() ([]byte, error) {
	return u, nil
}

func (u *UUID) UnmarshalText(text []byte) error {
	if text == nil {
		u = nil
		return nil
	}

	return u.Scan(text)
}

func (u UUID) String() string {
	return string(u)
}

func (u *UUID) Scan(src interface{}) error {
	b := src.([]byte)
	*u = make([]byte, len(b))
	copy(*u, b)
	return nil
}

func (u UUID) Value() (driver.Value, error) {
	return []byte(u), nil
}
