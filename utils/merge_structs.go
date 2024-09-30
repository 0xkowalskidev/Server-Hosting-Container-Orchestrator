package utils

import (
	"errors"
	"reflect"
)

// MergeStructs updates non-zero values from src into dst.
func MergeStructs(dst, src interface{}) error {
	// Ensure both src and dst are pointers
	dstVal := reflect.ValueOf(dst)
	srcVal := reflect.ValueOf(src)

	if dstVal.Kind() != reflect.Ptr || srcVal.Kind() != reflect.Ptr {
		return errors.New("both dst and src must be pointers")
	}

	// Dereference the pointers to get the underlying struct
	dstVal = dstVal.Elem()
	srcVal = srcVal.Elem()

	// Ensure both dst and src are structs
	if dstVal.Kind() != reflect.Struct || srcVal.Kind() != reflect.Struct {
		return errors.New("both dst and src must be structs")
	}

	// Iterate over the fields of the src struct and update dst with non-zero values
	for i := 0; i < srcVal.NumField(); i++ {
		srcField := srcVal.Field(i)
		dstField := dstVal.Field(i)

		// Check if the field is zero (empty) and only set if it's non-zero
		if !isZeroValue(srcField) {
			// Set the non-zero value from src into dst
			dstField.Set(srcField)
		}
	}

	return nil
}

// isZeroValue checks if a reflect.Value is zero (empty or default value)
func isZeroValue(v reflect.Value) bool {
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}
