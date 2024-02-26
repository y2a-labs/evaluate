package valid

import (
	"reflect"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// Global validator instance
var Validate *validator.Validate

func InitValidator() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

func ValidateStruct(ctx *fiber.Ctx, i interface{}) error {
	err := ctx.ParamsParser(i)
	if err != nil {
		return err
	}

	err = ctx.BodyParser(i)
	if err != nil {
		return err
	}

	SetDefaultValues(i)

	err = Validate.Struct(i)

	if err != nil {
		return err
	}

	return nil
}

func SetDefaultValues(v interface{}) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return
	}

	rv = rv.Elem()
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		if !field.CanSet() {
			continue
		}

		tag := rv.Type().Field(i).Tag.Get("default")
		if tag == "" {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			field.SetString(tag)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if value, err := strconv.ParseInt(tag, 10, 64); err == nil {
				field.SetInt(value)
			}
			// Add more cases here for other types as needed.
		}
	}
}
