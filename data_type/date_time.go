package data_type

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"auction-service/global"

	"github.com/go-playground/validator/v10"
)

type DateTime struct {
	origin   time.Time
	parseErr error
}

func (dt DateTime) layout() string {
	return time.RFC3339
}

func (dt DateTime) IsoLayout() string {
	return "YYYY-MM-DDThh:mm:ssZ"
}

func (dt *DateTime) parse(layout string, s string) {
	dt.origin, dt.parseErr = time.Time{}, nil
	if s == "" {
		return
	}

	t, err := time.Parse(layout, s)
	dt.parseErr = err
	if err == nil {
		*dt = NewDateTime(t)
	}
}

func (dt DateTime) HasParseErr() bool {
	return dt.parseErr != nil
}

func (dt DateTime) IsZero() bool {
	return dt.Time().IsZero()
}

func (dt DateTime) IsValid() bool {
	return !(dt.IsZero() || dt.HasParseErr())
}

func (dt DateTime) Format(layout string) string {
	return dt.Time().Format(layout)
}

func (dt DateTime) FormatUtc(layout string) string {
	return dt.Time().UTC().Format(layout)
}

func (dt DateTime) EndOfSecond() DateTime {
	return NewDateTime(
		dt.Time().
			Truncate(time.Second).
			Add(time.Second - time.Nanosecond),
	)
}

func (dt DateTime) NullDateTime() NullDateTime {
	return NewNullDateTime(&dt)
}

func (dt DateTime) Time() time.Time {
	return dt.origin
}

func (dt DateTime) Add(d time.Duration) DateTime {
	return NewDateTime(dt.Time().Add(d))
}

func (dt DateTime) Sub(u DateTime) time.Duration {
	return dt.Time().Sub(u.Time())
}

func (dt DateTime) IsEqual(u DateTime) bool {
	return dt.Time().Equal(u.Time())
}

func (dt DateTime) IsGreaterThan(u DateTime) bool {
	return dt.Time().After(u.Time())
}

func (dt DateTime) IsGreaterThanOrEqual(u DateTime) bool {
	return dt.IsGreaterThan(u) || dt.IsEqual(u)
}

func (dt DateTime) IsLessThan(u DateTime) bool {
	return dt.Time().Before(u.Time())
}

func (dt DateTime) IsLessThanOrEqual(u DateTime) bool {
	return dt.IsLessThan(u) || dt.IsEqual(u)
}

func (dt DateTime) String() string {
	l := dt.layout()
	s := "0000-00-00T00:00:00Z"

	if !dt.IsZero() {
		s = dt.Format(l)
	}

	return s
}

func (dt DateTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(dt.String())
}

func (dt *DateTime) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	dt.parse(dt.layout(), s)

	return nil
}

func (dt *DateTime) UnmarshalText(b []byte) error {
	dt.parse(dt.layout(), string(b))

	return nil
}

func (dt *DateTime) Scan(value interface{}) error {
	switch v := value.(type) {
	case time.Time:
		*dt = NewDateTime(v)
	default:
		return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type %T", value, dt)
	}

	return nil
}

func (dt DateTime) Value() (driver.Value, error) {
	l := "2006-01-02 15:04:05.999999"
	s := "0000-00-00 00:00:00.000000"

	if !dt.IsZero() {
		s = dt.FormatUtc(l)
	}

	return s, nil
}

func NewDateTime(t time.Time) DateTime {
	return DateTime{origin: t.In(global.GetTimeLocation())}
}

func DateTimeValidationFn(sl validator.StructLevel) {
	dateTime := sl.Current().Interface().(DateTime)
	if !dateTime.IsValid() {
		sl.ReportError(dateTime, "", "", "data_type_date_time", "")
	}
}
