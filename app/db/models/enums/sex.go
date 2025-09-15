package enums

import "database/sql/driver"

type Sex string

const (
	Female Sex = "female"
	Male       = "male"
)

func (s *Sex) Scan(value interface{}) error {
	*s = Sex(value.([]byte))
	return nil
}

func (s Sex) Value() (driver.Value, error) {
	return string(s), nil
}
