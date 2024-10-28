package responses

import (
	"fmt"
	"reflect"
	"time"
)

// ConvertTimeFieldsToUTC takes a pointer to a struct and converts all time.TimeString fields to UTC using reflection.
// Code is not be performant.
func ConvertTimeFieldsToUTC(v interface{}) {
	val := reflect.ValueOf(v)

	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		fmt.Println("Input must be a pointer to a struct")
		return
	}

	val = val.Elem()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if field.CanSet() && field.Kind() == reflect.Struct && field.Type() == reflect.TypeOf(time.Time{}) {
			t := field.Interface().(time.Time)
			field.Set(reflect.ValueOf(t.UTC()))
		}
	}
}
