package helper

import (
	"fmt"
	"reflect"
	"slices"
)

// data must be of type map[string]any, target must be a pointer to the desired output struct
func MapToStruct(data any, target any) []error {
	targetValue := reflect.ValueOf(target).Elem() // Dereference pointer to get actual struct
	targetType := targetValue.Type()              // Get Struct Type

	realData, ok := data.(map[string]any)
	if !ok {
		return []error{fmt.Errorf("data must be of type map[string]any")}
	}

	mapErrors := []error{}
	parsedMapKeys := []string{}

	for i := 0; i < targetValue.NumField(); i++ {
		fieldType := targetType.Field(i)

		// get map value according to struct field tag
		key := fieldType.Tag.Get("mapkey")
		if key == "" {
			mapErrors = append(mapErrors, fmt.Errorf(
				"target struct is missing tag 'mapkey' for field '%s'",
				fieldType.Name,
			))
			continue
		}
		val, ok := realData[key]
		if !ok {
			mapErrors = append(mapErrors, fmt.Errorf("key '%s' doesn't exist in data map", key))
			continue
		}

		// sanity check, that struct field can be set
		if !targetValue.Field(i).CanSet() {
			mapErrors = append(mapErrors, fmt.Errorf(
				"unable to set struct field '%s' (mapkey: %s) with value from map",
				fieldType.Name,
				key,
			))
			continue
		}

		// Check if map value is nil and set struct field accordingly
		if val == nil { //  && fieldType.Type.Kind() == reflect.Ptr
			targetValue.Field(i).Set(reflect.Zero(fieldType.Type))
			parsedMapKeys = append(parsedMapKeys, key)
			if fieldType.Type.Kind() != reflect.Ptr {
				mapErrors = append(mapErrors, fmt.Errorf(
					"map value for key '%s' is nil, but struct field '%s' is of non ptr type '%s', was assigned zero value anyway",
					key,
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

		// If map value isn't assignable to struct field, check if number conversion is needed
		if !valReflect.Type().AssignableTo(fieldType.Type) {
			if valReflect.Type().ConvertibleTo(targetType) {
				if validateConvertTypes(valReflect.Kind(), targetType.Kind()) {
					valReflect = valReflect.Convert(targetType)
				} else {
					mapErrors = append(mapErrors, fmt.Errorf(
						"map value of type '%s' for key '%s' is assignable to data type '%s' of struct field '%s' (target type %s), but no matching conversion found",
						valReflect.Type().String(),
						key,
						targetType.String(),
						fieldType.Name,
						targetType.Kind().String(),
					))
				}
			} else if valReflect.Kind() == reflect.Slice &&
				targetType.Kind() == reflect.Slice &&
				targetType.Elem().Kind() == reflect.String {

				valActual := valReflect.Interface().([]interface{})
				res := make([]string, len(valActual))
				for i, v := range valActual {
					res[i] = fmt.Sprintf("%v", v)
				}
				valReflect = reflect.ValueOf(res)
			} else if targetType.Kind() == reflect.Struct {
				mapErrors = append(mapErrors, fmt.Errorf("struct field '%s' is struct, map value is: %v", fieldType.Name, valReflect.Interface()))
				parsedMapKeys = append(parsedMapKeys, key)
				continue
				// TODO: make this recursive
			} else if targetType.Kind() == reflect.Slice &&
				targetType.Elem().Kind() == reflect.Struct {

				valActual := valReflect.Interface().([]interface{})
				for i, va := range valActual {
					mapErrors = append(mapErrors, fmt.Errorf("struct field '%s' is struct array, map value[%d] is: %v", fieldType.Name, i, va.(map[string]interface{})))
					// TODO: make this recursive
				}
				parsedMapKeys = append(parsedMapKeys, key)
				continue
			} else {
				mapErrors = append(mapErrors, fmt.Errorf(
					"map value of type '%s' for key '%s' is not assignable to data type '%s' of struct field '%s', ",
					valReflect.Type().String(),
					key,
					targetType.String(),
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
		targetValue.Field(i).Set(valReflect)
		parsedMapKeys = append(parsedMapKeys, key)
	}

	// Check if any map keys were missed
	for k := range realData {
		if !slices.Contains(parsedMapKeys, k) {
			mapErrors = append(mapErrors, fmt.Errorf("data map contains key '%s' which isn't present in target struct", k))
		}
	}
	return mapErrors
}

// TODO: move all this code to internal function which is called recursively on all child map[string]interface types
// func internalMapToStruct(data map[string]any, target reflect.Value, mapErrors *[]error) {
// 		// ...
// 		val, ok := data[key]
// 		if !ok {
// 			*mapErrors = append(*mapErrors, fmt.Errorf("key '%s' doesn't exist in data map", key))
// 			continue
// 		}
// 		// ...
// }

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
