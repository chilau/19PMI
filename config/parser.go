package config

import (
	"encoding/json"
	"github.com/hako/durafmt"
	"github.com/mcuadros/go-defaults"
	"reflect"
	"strconv"
	"strings"
)

func ParseConfig(data []byte) (config *ApplicationConfig, err error) {
	config = new(ApplicationConfig)
	defaults.SetDefaults(config)

	var configJson map[string]interface{}
	_ = json.Unmarshal(data, &configJson)

	parseConfigItem(config, configJson)

	return
}

func parseConfigItem(receiver interface{}, data map[string]interface{}) {
	if receiver == nil {
		return
	}

	receiverValue := reflect.ValueOf(receiver).Elem()
	for i := 0; i < receiverValue.NumField(); i++ {
		structField := receiverValue.Type().Field(i)
		fieldValue := receiverValue.FieldByName(structField.Name)

		jsonTag := getJsonTag(structField)
		if jsonTag != "" {
			value, found := data[jsonTag]
			if found {
				structData, isStruct := value.(map[string]interface{})
				arrayData, isArray := value.([]interface{})
				if !isStruct && !isArray {
					setValue(fieldValue, value)
				} else if isArray {
					arrayField := make([]interface{}, 0)
					for _, arrayItemData := range arrayData {
						arrayItemStructData, ok := arrayItemData.(map[string]interface{})
						if ok {
							arrayItem := createStruct(jsonTag)
							if arrayItem != nil {
								parseConfigItem(arrayItem, arrayItemStructData)
								arrayField = append(arrayField, arrayItem)
							}
						}
					}

					setValue(fieldValue, arrayField)
				} else {
					configItem := createStruct(jsonTag)
					if configItem == nil {
						continue
					}

					parseConfigItem(configItem, structData)
					setValue(fieldValue, configItem)
				}
			} else {
				requiredTag := getRequiredTag(structField)
				requiredFlag, err := strconv.ParseBool(requiredTag)
				if err == nil && requiredFlag {
					configItem := createStruct(jsonTag)
					setValue(fieldValue, configItem)
				}
			}
		}
	}

	return
}

func createStruct(fieldName string) (configItem interface{}) {
	switch fieldName {
	case "server":
		configItem = new(ServerConfig)
	}

	if configItem != nil {
		defaults.SetDefaults(configItem)
	}

	return
}

func setValue(
	fieldValue reflect.Value,
	value interface{},
) {
	if !fieldValue.IsValid() || !fieldValue.CanSet() {
		return
	}

	switch value.(type) {
	case string:
		timeout, err := durafmt.ParseString(value.(string))
		if err == nil {
			fieldValue.SetInt(int64(timeout.Duration()))
		} else {
			fieldValue.SetString(value.(string))
		}
	case bool:
		fieldValue.SetBool(value.(bool))
	case int:
		fieldValue.SetInt(value.(int64))
	case float64:
		f := value.(float64)
		fieldValue.SetInt(int64(f))
	default:
		fieldValue.Set(reflect.ValueOf(value))
	}

	return
}

func getJsonTag(field reflect.StructField) string {
	return getTag(field, "json")
}

func getRequiredTag(field reflect.StructField) string {
	return getTag(field, "required")
}

func getTag(field reflect.StructField, tagName string) (tagValue string) {
	varName := field.Name

	tag := field.Tag.Get(tagName)
	if tag == "" || tag == "-" {
		return
	}

	parts := strings.Split(tag, ",")
	value := parts[0]
	if value == "" {
		value = varName
	}

	tagValue = value

	return
}
