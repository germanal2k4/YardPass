package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func processValue(v reflect.Value) error {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if !field.CanSet() {
			continue
		}

		if field.Kind() == reflect.Struct && fieldType.Type != reflect.TypeOf(time.Duration(0)) {
			if err := processValue(field); err != nil {
				return err
			}
			continue
		}

		envKey := fieldType.Tag.Get("env")
		defaultVal := fieldType.Tag.Get("default")

		if isZero(field) && defaultVal != "" {
			if err := setFieldValue(field, defaultVal); err != nil {
				return fmt.Errorf("set default for %s: %w", fieldType.Name, err)
			}
		}

		if envKey != "" {
			if envVal := os.Getenv(envKey); envVal != "" {
				if err := setFieldValue(field, envVal); err != nil {
					return fmt.Errorf("set env %s for %s: %w", envKey, fieldType.Name, err)
				}
			}
		}
	}

	return nil
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Slice, reflect.Map:
		return v.IsNil() || v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	default:
		return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
	}
}

func buildDefaultsMap(t reflect.Type, v reflect.Value) map[string]any {
	result := make(map[string]any)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		yamlTag := field.Tag.Get("yaml")
		if yamlTag == "" || yamlTag == "-" {
			continue
		}

		yamlKey := strings.Split(yamlTag, ",")[0]

		if field.Type.Kind() == reflect.Struct && field.Type != reflect.TypeOf(time.Duration(0)) {
			result[yamlKey] = buildDefaultsMap(field.Type, fieldValue)
			continue
		}

		if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct {
			// Create zero value of the pointed type
			elemType := field.Type.Elem()
			result[yamlKey] = buildDefaultsMap(elemType, reflect.Zero(elemType))
			continue
		}

		defaultVal := field.Tag.Get("default")
		if defaultVal == "" {
			continue
		}

		result[yamlKey] = convertDefaultValue(field.Type, defaultVal)
	}

	return result
}

func convertDefaultValue(t reflect.Type, value string) any {
	switch t.Kind() {
	case reflect.String:
		return value
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if t == reflect.TypeOf(time.Duration(0)) {
			return value
		}
		if i, err := strconv.ParseInt(value, 10, 64); err == nil {
			return i
		}
		return value
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if u, err := strconv.ParseUint(value, 10, 64); err == nil {
			return u
		}
		return value
	case reflect.Float32, reflect.Float64:
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
		return value
	case reflect.Bool:
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
		return value
	case reflect.Slice:
		if t.Elem().Kind() == reflect.String {
			parts := strings.Split(value, ",")
			for i := range parts {
				parts[i] = strings.TrimSpace(parts[i])
			}
			return parts
		}
		return value
	default:
		return value
	}
}

func mergeMaps(base, override map[string]any) map[string]any {
	result := make(map[string]any)

	for k, v := range base {
		result[k] = v
	}

	for k, v := range override {
		if baseVal, exists := result[k]; exists {
			baseMap, baseIsMap := baseVal.(map[string]any)
			overrideMap, overrideIsMap := v.(map[string]any)
			if baseIsMap && overrideIsMap {
				result[k] = mergeMaps(baseMap, overrideMap)
				continue
			}
		}
		result[k] = v
	}

	return result
}

func setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			d, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("parse duration %q: %w", value, err)
			}
			field.SetInt(int64(d))
			return nil
		}
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("parse int %q: %w", value, err)
		}
		field.SetInt(i)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("parse uint %q: %w", value, err)
		}
		field.SetUint(u)

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("parse float %q: %w", value, err)
		}
		field.SetFloat(f)

	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("parse bool %q: %w", value, err)
		}
		field.SetBool(b)

	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.String {
			parts := strings.Split(value, ",")
			for i := range parts {
				parts[i] = strings.TrimSpace(parts[i])
			}
			field.Set(reflect.ValueOf(parts))
		} else {
			return fmt.Errorf("unsupported slice type: %s", field.Type())
		}

	default:
		return fmt.Errorf("unsupported type: %s", field.Kind())
	}

	return nil
}
