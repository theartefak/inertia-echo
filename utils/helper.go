package utils

import (
	"os"
	"reflect"
)

// MergeMaps merges two maps, overriding existing keys from b into a new map.
func MergeMaps(a, b map[string]interface{}) map[string]interface{} {
	retVal := make(map[string]interface{}, len(a)+len(b))
	for k, v := range a {
		retVal[k] = v
	}
	for k, v := range b {
		retVal[k] = v
	}
	return retVal
}

// InArray checks if a given value exists in the given array.
func InArray(val interface{}, array interface{}) (exists bool, index int) {
	exists = false
	index = -1

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) {
				index = i
				exists = true
				return
			}
		}
	}

	return
}

// WalkRecursive iterates over a given map recursively and calls the given func on every value.
func WalkRecursive(values map[string]interface{}, fn func(interface{})) {
	for _, v := range values {
		switch v := v.(type) {
		case map[string]interface{}:
			WalkRecursive(v, fn)
		default:
			fn(v)
		}
	}
}

// GetEnvOrDefault gets an environment variable or returns the default value.
func GetEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultValue
}
