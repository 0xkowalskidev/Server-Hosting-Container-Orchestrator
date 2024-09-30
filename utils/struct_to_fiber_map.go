package utils

import (
	"github.com/gofiber/fiber/v3"
	"reflect"
)

func StructToFiberMap(input interface{}) fiber.Map {
	result := fiber.Map{}

	// Ensure the input is a pointer or struct
	val := reflect.ValueOf(input)
	if val.Kind() == reflect.Ptr {
		val = val.Elem() // Dereference pointer
	}

	// Ensure we are working with a struct
	if val.Kind() != reflect.Struct {
		return result
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		if !field.IsExported() {
			continue
		}

		result[field.Name] = fieldValue.Interface()
	}

	return result
}
