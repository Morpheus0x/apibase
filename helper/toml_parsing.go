package helper

import (
	"reflect"
	"time"

	"gopkg.cc/apibase/errx"
	"gopkg.cc/apibase/log"
)

// This function parses any struct field with a 'toml' tag from config into the field with tag 'internal' of the same name, if present.
// If value of struct field with 'toml' tag is unset (meaning it has default zero value), the field with tag 'internal' of same name of struct defaults is used,
// or if no 'internal' tag of same name exists, the value of the field with 'toml' tag from defaults is used.
// Alongside the "internal" tag there may exist a tag called "parsetype" which processes the config value in a special way:
//
// parsetype:"percentage" - parses a percentage string (e.g. "20%") from field with 'toml' tag to float32 in field with 'internal' tag
func ParseTomlConfigAndDefaults[T any](config *T, defaults *T) error {
	if reflect.TypeOf(defaults).Kind() != reflect.Ptr ||
		reflect.ValueOf(defaults).Elem().Kind() != reflect.Struct {
		return errx.New("defaults must be a pointer to a struct")
	}
	if reflect.TypeOf(config).Kind() != reflect.Ptr ||
		reflect.TypeOf(config).Elem().Kind() != reflect.Struct {
		return errx.New("config must be a pointer to a struct")
	}
	// TODO: type comparison is redundant, because it is enfocred by the generic function
	if reflect.ValueOf(config).Elem().Type() != reflect.ValueOf(defaults).Elem().Type() {
		return errx.New("config and defaults must be of same type")
	}

	internalTags := GetIndexForTag(config, "internal")
	configStruct := reflect.ValueOf(config).Elem()
	defaultsStruct := reflect.ValueOf(defaults).Elem()
	for i := 0; i < configStruct.NumField(); i++ {
		tomlTag, tomlTagExists := configStruct.Type().Field(i).Tag.Lookup("toml")
		dataIndex, ok := internalTags[tomlTag]
		if !ok {
			// if no "internal" tag with same content as "toml" tag exists, use value with "toml" tag from defaults
			dataIndex = i
		}
		if _, ok := configStruct.Type().Field(i).Tag.Lookup("internal"); ok || !tomlTagExists {
			continue
		}
		if !configStruct.Field(dataIndex).CanSet() {
			return errx.Newf("Unable to set internal data struct field at index %d", dataIndex)
		}

		configString, isConfigString := configStruct.Field(i).Interface().(string)
		configFieldType := configStruct.Field(dataIndex).Type()
		parseType, hasParseType := configStruct.Type().Field(dataIndex).Tag.Lookup("parsetype")
		if configFieldType == reflect.TypeOf(time.Duration(0)) && isConfigString {
			duration, err := StringToDuration(configString)
			if err != nil || configStruct.Field(i).IsZero() {
				log.Logf(log.LevelWarning,
					"Unable to parse field with toml tag %s of struct %s: '%v', assuming default '%s'",
					tomlTag,
					configStruct.Type().String(),
					configStruct.Field(i).Interface(),
					defaultsStruct.Field(dataIndex).Interface().(time.Duration).String(),
				)
				configStruct.Field(dataIndex).Set(defaultsStruct.Field(dataIndex))
				continue
			}
			configStruct.Field(dataIndex).Set(reflect.ValueOf(duration))
		} else if configFieldType == reflect.TypeOf(float32(0)) && isConfigString && hasParseType && parseType == "percentage" {
			percentage, err := PercentageToFloat32(configString)
			if err != nil || configStruct.Field(i).IsZero() {
				log.Logf(log.LevelWarning,
					"Unable to parse field with toml tag %s of struct %s: '%v', assuming default '%s%%'",
					tomlTag,
					configStruct.Type().String(),
					configStruct.Field(i).Interface(),
					FancyFloat(defaultsStruct.Field(dataIndex).Interface().(float64)*100),
				)
				configStruct.Field(dataIndex).Set(defaultsStruct.Field(dataIndex))
				continue
			}
			configStruct.Field(dataIndex).Set(reflect.ValueOf(percentage))
		} else if configStruct.Field(i).IsZero() {
			log.Logf(log.LevelWarning,
				"Unable to parse field with toml tag %s of struct %s: '%v', assuming default '%v'",
				tomlTag,
				configStruct.Type().String(),
				configStruct.Field(i).Interface(),
				defaultsStruct.Field(dataIndex).Interface(),
			)
			configStruct.Field(dataIndex).Set(defaultsStruct.Field(dataIndex))
		} else {
			// Nothing to do since parsed value from "toml" tag is of desired type
		}
	}
	return nil
}

func GetIndexForTag(data any, tag string) map[string]int {
	result := make(map[string]int)
	dataType := reflect.TypeOf(data)
	if dataType.Kind() == reflect.Ptr {
		dataType = dataType.Elem()
	}
	if dataType.Kind() != reflect.Struct {
		return result
	}
	for i := 0; i < dataType.NumField(); i++ {
		if tag, ok := dataType.Field(i).Tag.Lookup(tag); ok {
			result[tag] = i
		}
	}
	return result
}
