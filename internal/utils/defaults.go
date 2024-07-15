package utils

import (
	"errors"
	"reflect"
	"strconv"
)

func ApplyDefaults(s any) error {
	val := reflect.ValueOf(s)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return errors.New("pointer to a struct is required")
	}

	return processFields(val.Elem())
}

func processFields(val reflect.Value) error {
	for i := 0; i < val.NumField(); i++ {
		err := processField(val.Field(i), val.Type().Field(i))
		if err != nil {
			return err
		}
	}

	return nil
}

func processField(val reflect.Value, sfield reflect.StructField) error {
	defVal, defExist := sfield.Tag.Lookup("default")
	kind := val.Kind()
	isStruct := kind == reflect.Struct || (kind == reflect.Ptr && val.Elem().Kind() == reflect.Struct)
	if (!val.IsZero() || !defExist) && !isStruct {
		return nil
	}
	switch kind {
	case reflect.String:
		val.SetString(defVal)
	case reflect.Bool:
		b, err := strconv.ParseBool(defVal)
		if err != nil {
			return err
		}
		val.SetBool(b)
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		i, err := strconv.ParseInt(defVal, 10, int(val.Type().Size())*8)
		if err != nil {
			return err
		}
		val.SetInt(i)
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		u, err := strconv.ParseUint(defVal, 10, int(val.Type().Size())*8)
		if err != nil {
			return err
		}
		val.SetUint(u)
	case reflect.Struct:
		return processFields(val)
	case reflect.Ptr:
		elem := val.Elem()
		switch elem.Kind() {
		case reflect.Struct:
			if !val.IsNil() {
				return processFields(elem)
			}
		default:
			return errors.New("unsupported field type")
		}
	default:
		return errors.New("unsupported field type")
	}
	return nil
}
