package enums

import "database/sql/driver"

type Doctype string

const (
	Rules       Doctype = "rules"
	Regulations         = "regulations"
)

func (d *Doctype) Scan(value interface{}) error {
	*d = Doctype(value.([]byte))
	return nil
}

func (d Doctype) Value() (driver.Value, error) {
	return string(d), nil
}
