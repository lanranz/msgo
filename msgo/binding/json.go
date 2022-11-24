package binding

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
)

type jsonBinding struct {
	DisallowUnknownFields bool
	IsValidate            bool
}

var Validator StructValidator = &defaultValidator{}

func (b jsonBinding) Name() string {
	return "json"
}

func (b jsonBinding) Bind(r *http.Request, obj any) error {
	body := r.Body
	if body == nil {
		return errors.New("invalid request")
	}
	decoder := json.NewDecoder(body) //解码器
	if b.DisallowUnknownFields {
		decoder.DisallowUnknownFields() //如果有未知的字段会报错
	}
	if b.IsValidate {
		err := ValidateParam(obj, decoder)
		if err != nil {
			return err
		}
	} else {
		err := decoder.Decode(obj)
		if err != nil {
			return err
		}
	}
	return validate(obj)
}

//func validateStruct(obj any) error {
//	return validator.New().Struct(obj)
//}

func ValidateParam(obj any, decoder *json.Decoder) error {
	//解析为map，根据map中的key进行比对
	//判断类型 结构体类型时才能机型判断
	//判断那类型用到反射
	valueOf := reflect.ValueOf(obj)
	//判断是否是指针类型
	if valueOf.Kind() != reflect.Pointer {
		return errors.New("no ptr type")
	}
	elem := valueOf.Elem().Interface()
	of := reflect.ValueOf(elem)

	switch of.Kind() {
	case reflect.Struct:
		return cheackParam(of, obj, decoder)
	case reflect.Slice, reflect.Array:
		elem := of.Type().Elem()
		if elem.Kind() == reflect.Struct {
			return cheackSlice(elem, obj, decoder)
		}
	default:
		_ = decoder.Decode(obj)
	}
	return nil
}

func cheackSlice(of reflect.Type, obj any, decoder *json.Decoder) error {
	mapValue := make([]map[string]interface{}, 0)
	_ = decoder.Decode(&mapValue)
	for i := 0; i < of.NumField(); i++ {
		filed := of.Field(i)
		name := filed.Name
		jsonName := filed.Tag.Get("json")
		if jsonName != "" {
			name = jsonName
		}
		required := filed.Tag.Get("msgo")
		for _, v := range mapValue {
			value := v[name]
			if value == nil && required == "required" {
				return errors.New(fmt.Sprintf("filed [%s] is not exist,because [%s] is required", jsonName, jsonName))
			}
		}
	}
	b, _ := json.Marshal(mapValue)
	_ = json.Unmarshal(b, obj)
	return nil
}

func cheackParam(of reflect.Value, obj any, decoder *json.Decoder) error {
	//解析为map，并通过map形式比对
	//判断类型为结构体时才能解析为map
	mapValue := make(map[string]interface{})
	_ = decoder.Decode(&mapValue)
	for i := 0; i < of.NumField(); i++ {
		filed := of.Type().Field(i)
		name := filed.Name
		jsonName := filed.Tag.Get("json")
		if jsonName != "" {
			name = jsonName
		}
		required := filed.Tag.Get("msgo")
		value := mapValue[name]
		if value == nil && required == "required" {
			return errors.New(fmt.Sprintf("filed [%s] is not exist,because [%s] is required", jsonName, jsonName))
		}
	}
	b, _ := json.Marshal(mapValue)
	_ = json.Unmarshal(b, obj)
	return nil
}
