package data_type

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"auction-service/constant"

	"github.com/go-playground/validator/v10"
)

type NullDateTime struct {
	dateTime DateTime
	isValid  bool
}

func (dt NullDateTime) layout() string {
	return dt.dateTime.layout()
}

func (dt NullDateTime) IsoLayout() string {
	return dt.dateTime.IsoLayout()
}

func (dt NullDateTime) get() *DateTime {
	var dateTime *DateTime

	if dt.isValid {
		dateTime = new(DateTime)
		*dateTime = dt.dateTime
	}

	return dateTime
}

func (dt *NullDateTime) set(v *DateTime) {
	dateTime, isValid := DateTime{}, false
	if v != nil {
		dateTime, isValid = *v, true
	}

	dt.dateTime, dt.isValid = dateTime, isValid
}

func (dt *NullDateTime) parse(s *string) {
	if s == nil || *s == constant.NilAsString {
		dt.set(nil)
		return
	}

	dateTime := DateTime{}
	dateTime.parse(dt.layout(), *s)
	dt.set(&dateTime)
}

func (dt NullDateTime) IsNil() bool {
	return dt.get() == nil
}

func (dt NullDateTime) DateTime() DateTime {
	dateTime := dt.get()
	if dateTime == nil {
		return DateTime{}
	}

	return *dateTime
}

func (dt NullDateTime) DateTimeP() *DateTime {
	return dt.get()
}

func (dt NullDateTime) Add(d time.Duration) NullDateTime {
	dateTime := dt.get()
	if dateTime != nil {
		*dateTime = dateTime.Add(d)
	}

	return NewNullDateTime(dateTime)
}

func (dt NullDateTime) IsEqual(u NullDateTime) bool {
	value, operand := dt.DateTimeP(), u.DateTimeP()
	if (value == nil && operand != nil) || (operand == nil && value != nil) {
		return false
	}

	return (value == nil && operand == nil) || (value.IsEqual(*operand))
}

func (dt NullDateTime) IsGreaterThan(u NullDateTime) bool {
	value, operand := dt.DateTimeP(), u.DateTimeP()

	// nil is greater than anything
	isGreaterThan := value == nil
	if value != nil && operand != nil {
		isGreaterThan = value.IsGreaterThan(*operand)
	}

	return isGreaterThan
}

func (dt NullDateTime) IsGreaterThanOrEqual(u NullDateTime) bool {
	return dt.IsGreaterThan(u) || dt.IsEqual(u)
}

func (dt NullDateTime) IsLessThan(u NullDateTime) bool {
	value, operand := dt.DateTimeP(), u.DateTimeP()

	// nil is greater than anything
	isLessThan := value != nil
	if value != nil && operand != nil {
		isLessThan = value.IsLessThan(*operand)
	}

	return isLessThan
}

func (dt NullDateTime) IsLessThanOrEqual(u NullDateTime) bool {
	return dt.IsLessThan(u) || dt.IsEqual(u)
}

func (dt NullDateTime) String() string {
	dateTime := dt.get()
	if dateTime == nil {
		return constant.NilAsString
	}

	return dateTime.String()
}

func (dt NullDateTime) MarshalJSON() ([]byte, error) {
	dateTime := dt.get()
	if dateTime == nil {
		return []byte(`null`), nil
	}

	return json.Marshal(dateTime.String())
}

func (dt *NullDateTime) UnmarshalJSON(b []byte) error {
	var s *string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	dt.parse(s)

	return nil
}

func (dt *NullDateTime) UnmarshalText(b []byte) error {
	s := string(b)
	dt.parse(&s)

	return nil
}

func (dt *NullDateTime) Scan(value interface{}) error {
	if value == nil {
		dt.set(nil)
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		dateTime := NewDateTime(v)
		dt.set(&dateTime)
	default:
		return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type %T", value, dt)
	}

	return nil
}

func (dt NullDateTime) Value() (driver.Value, error) {
	dateTime := dt.get()
	if dateTime == nil {
		return nil, nil
	}

	return dateTime.Value()
}

func NewNullDateTime(v *DateTime) NullDateTime {
	dt := NullDateTime{}
	dt.set(v)

	return dt
}

func NullDateTimeValidationFn(sl validator.StructLevel) {
	nullDateTime := sl.Current().Interface().(NullDateTime)
	dateTime := nullDateTime.DateTimeP()
	if dateTime != nil && !dateTime.IsValid() {
		sl.ReportError(nullDateTime, "", "", "data_type_null_date_time", "")
	}
}
