package cmdopts

import (
	"fmt"
	"reflect"
	"strings"
)

// Options represents map or struct (or pointer to struct as a sequence of long named perameters
func Options(src interface{}) ([]string, error) {
	return generateOptions(src, src)
}

func generateOptions(orig interface{}, src interface{}) ([]string, error) {
	value := reflect.ValueOf(src)
	switch t := reflect.TypeOf(src); t.Kind() {
	case reflect.Map:
		if t.Key().Kind() != reflect.String {
			return nil, fmt.Errorf("only map[string]T is accepted, got %T", orig)
		}
		var res []string
		for _, key := range value.MapKeys() {
			v := value.MapIndex(key)
			res = append(res, "--"+key.String(), fmt.Sprintf("%v", v.Interface()))
		}
		return res, nil
	case reflect.Struct:
		var res []string
		for i := 0; i < t.NumField(); i++ {
			field := value.Field(i)
			fieldType := t.Field(i)
			if fieldType.Anonymous {
				// анонимная хрень, ныряем вниз
				sub, err := generateOptions(field.Interface(), field.Interface())
				if err != nil {
					return nil, err
				}
				res = append(res, sub...)
				continue
			}

			if !isPublic(fieldType.Name) {
				continue
			}
			paramName, ok := fieldType.Tag.Lookup("param")
			if !ok {
				return nil, fmt.Errorf("public field %s must have param tag which is to be used for param name", fieldType.Name)
			}

			res = append(res, "--"+paramName, fmt.Sprintf("%v", field.Interface()))
		}
		return res, nil
	case reflect.Ptr:
		return generateOptions(orig, value.Elem().Interface())
	default:
		return nil, fmt.Errorf("map[string]T or struct (pointer to struct) expected, got %T", orig)
	}
}

func isPublic(s string) bool {
	if len(s) == 0 {
		return false
	}
	return strings.ToUpper(s)[0] == s[0]
}
