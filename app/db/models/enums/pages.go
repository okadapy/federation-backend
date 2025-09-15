package enums

import "database/sql/driver"

type Page string

const (
	News      Page = "news"
	Gallery        = "gallery"
	Documents      = "documents"
)

func (p *Page) Scan(value interface{}) error {
	*p = Page(value.([]byte))
	return nil
}

func (p Page) Value() (driver.Value, error) {
	return string(p), nil
}
