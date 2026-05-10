package shared

import (
	"strings"
	"time"
)

const dateLayout = "2006-01-02"

// Date é um time.Time que serializa/deserializa JSON como "YYYY-MM-DD".
type Date struct{ time.Time }

func (d *Date) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "null" || s == "" {
		return nil
	}
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}

func (d Date) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + d.Format(dateLayout) + `"`), nil
}
