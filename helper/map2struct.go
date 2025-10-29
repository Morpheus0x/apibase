package helper

import (
	"fmt"
	"reflect"
	"slices"

	"gopkg.cc/apibase/errx"
)

var (
	ErrM2S_Critical = errx.NewType("unable to convert")
	ErrM2S_Mismatch = errx.NewType("map to struct mismatch")
	ErrM2S_Warning  = errx.NewType("warning")
)

// data must be of type map[string]any,
// target must be a pointer to the desired output struct,
// the target struct must have tag `mapkey` corresponding to the map key being parsed
func MapToStruct(data any, target any) []error {
	targetValue := reflect.ValueOf(target).Elem() // Dereference pointer to get actual struct

	realData, ok := data.(map[string]any)
	if !ok {
		return []error{errx.NewWithTypef(ErrM2S_Critical, "data must be of type map[string]any")}
	}

	mapErrors := []error{}
	internalMapToStruct(realData, targetValue, &mapErrors, "", "")
	return mapErrors
}

func internalMapToStruct(data map[string]any, target reflect.Value, mapErrors *[]error, parentStructField string, parentMapKey string) {
	targetType := target.Type() // Get Struct Type
	parsedMapKeys := []string{}

	for i := 0; i < target.NumField(); i++ {
		fieldType := targetType.Field(i)

		// get map value according to struct field tag
		key := fieldType.Tag.Get("mapkey")
		if key == "" {
			*mapErrors = append(*mapErrors, errx.NewWithTypef(ErrM2S_Critical,
				"target struct is missing tag 'mapkey' for field '%s%s'",
				parentStructField,
				fieldType.Name,
			))
			continue
		}
		val, ok := data[key]
		if !ok {
			*mapErrors = append(*mapErrors, errx.NewWithTypef(ErrM2S_Mismatch, "key '%s%s' doesn't exist in data map", parentMapKey, key))
			continue
		}

		// sanity check, that struct field can be set
		if !target.Field(i).CanSet() {
			*mapErrors = append(*mapErrors, errx.NewWithTypef(ErrM2S_Critical,
				"unable to set struct field '%s%s' (mapkey: %s) with value from map",
				parentStructField,
				fieldType.Name,
				key,
			))
			continue
		}

		// Check if map value is nil and set struct field accordingly
		if val == nil { //  && fieldType.Type.Kind() == reflect.Ptr
			target.Field(i).Set(reflect.Zero(fieldType.Type))
			parsedMapKeys = append(parsedMapKeys, key)
			if fieldType.Type.Kind() != reflect.Ptr {
				*mapErrors = append(*mapErrors, errx.NewWithTypef(ErrM2S_Mismatch,
					"map value for key '%s%s' is nil, but struct field '%s%s' is of non ptr type '%s', was assigned zero value anyway",
					parentMapKey,
					key,
					parentStructField,
					fieldType.Name,
					fieldType.Type.String(),
				))
			}
			parsedMapKeys = append(parsedMapKeys, key)
			continue
		}

		// Check if map value is ptr and dereference it
		valReflect := reflect.ValueOf(val)
		if reflect.TypeOf(val).Kind() == reflect.Ptr {
			valReflect = valReflect.Elem()
		}

		// Check if target field is ptr and get underlying type
		targetType := fieldType.Type
		if fieldType.Type.Kind() == reflect.Ptr {
			targetType = fieldType.Type.Elem()
		}

		// If map value isn't directly assignable to struct field, check for valid conversions
		if !valReflect.Type().AssignableTo(targetType) { // TODO: check if AssignableTo(targetType) or AssignableTo(fieldType.Type) is correct
			if valReflect.Type().ConvertibleTo(targetType) {
				// Process any number to target struct field type
				if validateConvertTypes(valReflect.Kind(), targetType.Kind()) {
					valReflect = valReflect.Convert(targetType)
				} else {
					*mapErrors = append(*mapErrors, errx.NewWithTypef(ErrM2S_Critical,
						"map value of type '%s' for key '%s%s' is assignable to data type '%s' of struct field '%s%s' (target type %s), but no matching conversion found",
						valReflect.Type().String(),
						parentMapKey,
						key,
						targetType.String(),
						parentStructField,
						fieldType.Name,
						targetType.Kind().String(),
					))
				}
			} else if valReflect.Kind() == reflect.Slice &&
				targetType.Kind() == reflect.Slice &&
				targetType.Elem().Kind() == reflect.String {
				// Process []string
				// TODO: implement processing of any array of convertible values
				valActual := valReflect.Interface().([]interface{})
				res := make([]string, len(valActual))
				for i, v := range valActual {
					res[i] = fmt.Sprintf("%v", v)
				}
				valReflect = reflect.ValueOf(res)
			} else if targetType.Kind() == reflect.Struct {
				// Recursively process nested struct
				valReflect = reflect.New(targetType).Elem()
				internalMapToStruct(
					valReflect.Interface().(map[string]interface{}),
					valReflect,
					mapErrors,
					getParentFieldKey(fieldType.Name, parentStructField, -1),
					getParentFieldKey(key, parentMapKey, -1),
				)
			} else if targetType.Kind() == reflect.Slice &&
				targetType.Elem().Kind() == reflect.Struct {
				// Recursively process nested []struct
				valActual := valReflect.Interface().([]interface{})
				valReflect = reflect.MakeSlice(targetType, len(valActual), len(valActual))
				for k, va := range valActual {
					elemValReflect := reflect.New(targetType.Elem()).Elem()
					internalMapToStruct(
						va.(map[string]interface{}),
						elemValReflect,
						mapErrors,
						getParentFieldKey(fieldType.Name, parentStructField, k),
						getParentFieldKey(key, parentMapKey, k),
					)
					valReflect.Index(k).Set(elemValReflect)
				}
			} else {
				*mapErrors = append(*mapErrors, errx.NewWithTypef(ErrM2S_Critical,
					"map value of type '%s' for key '%s%s' is not assignable to data type '%s' of struct field '%s%s', ",
					valReflect.Type().String(),
					parentMapKey,
					key,
					targetType.String(),
					parentStructField,
					fieldType.Name,
				))
				parsedMapKeys = append(parsedMapKeys, key)
				continue
			}
		}

		// If target is ptr, convert map value to ptr
		if fieldType.Type.Kind() == reflect.Ptr {
			ptr := reflect.New(targetType)
			ptr.Elem().Set(valReflect)
			valReflect = ptr
		}

		// Set target struct field
		target.Field(i).Set(valReflect)
		parsedMapKeys = append(parsedMapKeys, key)
	}

	// Check if any map keys were missed
	for k := range data {
		if !slices.Contains(parsedMapKeys, k) {
			*mapErrors = append(*mapErrors, errx.NewWithTypef(ErrM2S_Warning,
				"data map contains key '%s%s' which isn't present in target struct",
				parentMapKey,
				k,
			))
		}
	}
}

func validateConvertTypes(k reflect.Kind, l reflect.Kind) bool {
	if k == reflect.String && l == reflect.String {
		return true
	}
	if k == reflect.Bool && l == reflect.Bool {
		return true
	}
	if (k >= reflect.Int && k <= reflect.Float64) && (l >= reflect.Int && l <= reflect.Float64) {
		return true
	}
	return false
}

func getParentFieldKey(name string, parent string, i int) string {
	index := ""
	if i != -1 {
		index = fmt.Sprintf("[%d]", i)
	}
	if parent == "" {
		return fmt.Sprintf("%s%s.", name, index)
	} else {
		return fmt.Sprintf("%s%s%s.", parent, name, index)
	}
}
