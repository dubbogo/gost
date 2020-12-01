/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package json_struct_parser

import (
	"fmt"
	"io/ioutil"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	hessian2 "github.com/apache/dubbo-go-hessian2"
	"github.com/buger/jsonparser"
)

type jsonStructParser struct {
	structFields   []reflect.StructField
	valueMap       map[string]string
	subObjValueMap map[string]reflect.Value
}

func newJsonStructParser() *jsonStructParser {
	return &jsonStructParser{
		structFields:   make([]reflect.StructField, 0),
		valueMap:       make(map[string]string),
		subObjValueMap: make(map[string]reflect.Value),
	}
}

func init() {
	defaultJsonStructParser = newJsonStructParser()
}

var defaultJsonStructParser *jsonStructParser

func JsonFile2Interface(path string) (interface{}, error) {
	defer func() {
		defaultJsonStructParser = newJsonStructParser()
	}()
	return defaultJsonStructParser.Jsonfile2Struct(path)
}

func RemoveTargetNameField(v interface{}, targetName string) interface{} {
	defer func() {
		defaultJsonStructParser = newJsonStructParser()
	}()
	return defaultJsonStructParser.RemoveTargetNameField(v, targetName)
}

func (jsp *jsonStructParser) cb(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
	switch dataType {
	case jsonparser.Object: // 嵌套子结构
		newParser := newJsonStructParser()
		subObj := newParser.json2Struct(value)
		hessian2.RegisterPOJOMapping(getJavaClassName(subObj), subObj)
		jsp.structFields = append(jsp.structFields, reflect.StructField{
			Name: string(key),
			Type: reflect.TypeOf(subObj),
		})
		jsp.subObjValueMap[string(key)] = reflect.ValueOf(subObj)
	case jsonparser.Array: //数组结构
		newParser := newJsonStructParser()
		subObj := newParser.json2Struct(value)
		hessian2.RegisterPOJOMapping(getJavaClassName(subObj), subObj) // TODO 目前存在问题
		jsp.structFields = append(jsp.structFields, reflect.StructField{
			Name: string(key),
			Type: reflect.TypeOf(subObj),
		})

	case jsonparser.String: // 正常结构
		// "type@value"
		arr := strings.Split(string(value), "@")
		var userDefinedType reflect.Type
		switch arr[0] {
		case "int":
			userDefinedType = reflect.TypeOf(0)
		case "string":
			userDefinedType = reflect.TypeOf("")
		case "uint64":
			userDefinedType = reflect.TypeOf(uint64(0))
		case "time.Time":
			userDefinedType = reflect.TypeOf(time.Time{})
		case "float32":
			userDefinedType = reflect.TypeOf(float32(0))
		case "float64":
			userDefinedType = reflect.TypeOf(float64(0))
		case "bool":
			userDefinedType = reflect.TypeOf(false)
		default:
			log.Println("warning: val", string(value), " in json is not supported")
		}
		if len(arr) > 1 {
			jsp.valueMap[string(key)] = arr[1]
		}
		jsp.structFields = append(jsp.structFields, reflect.StructField{
			Name: string(key),
			Type: userDefinedType,
		})
	default:
		log.Println("warning: dataType ", string(value), " in json is not supported")
	}
	return nil
}

func (jsp *jsonStructParser) json2Struct(jsonData []byte) interface{} {
	if err := jsonparser.ObjectEach(jsonData, jsp.cb); err != nil {
		log.Println("jsonparser.ObjectEach error = ", err)
	}
	typ := reflect.StructOf(jsp.structFields)
	v := reflect.New(typ).Elem()
	newty := reflect.TypeOf(v.Addr().Interface()).Elem()
	for i := 0; i < typ.NumField(); i++ {
		valStr, ok1 := jsp.valueMap[newty.Field(i).Name]
		subObj, ok2 := jsp.subObjValueMap[newty.Field(i).Name]
		if !ok1 && !ok2 {
			continue
		}

		if newty.Field(i).Type.Kind() == reflect.Ptr {
			v.Field(i).Set(subObj)
			continue
		}
		switch newty.Field(i).Type {
		case reflect.TypeOf(0), reflect.TypeOf(uint64(0)):
			if parsedInt, err := strconv.Atoi(valStr); err == nil {
				v.Field(i).SetInt(int64(parsedInt))
				break
			}
			v.Field(i).SetInt(0)
		case reflect.TypeOf(""):
			v.Field(i).SetString(valStr)
		case reflect.TypeOf(time.Time{}):
			//todo time support v.Field(i).
		case reflect.TypeOf(float64(0)), reflect.TypeOf(float32(0)):
			if parsedFloat, err := strconv.ParseFloat(valStr, 64); err == nil {
				v.Field(i).SetFloat(parsedFloat)
				break
			}
			v.Field(i).SetFloat(0)
		case reflect.TypeOf(false):
			if valStr == "true" || valStr == "1" {
				v.Field(i).SetBool(true)
			}

		default:
			log.Println("warning val: ", valStr, " in value is not supported")
		}
	}
	s := v.Addr().Interface()
	return s
}

// Jsonfile2Struct read file from @filePath and parse data to interface
func (jsp *jsonStructParser) Jsonfile2Struct(filePath string) (interface{}, error) {
	jsonData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return jsp.json2Struct(jsonData), nil
}

// RemoveTargetNameField remove origin interface @v's target field by @targetName
func (jsp *jsonStructParser) RemoveTargetNameField(v interface{}, targetName string) interface{} {
	typ := reflect.TypeOf(v).Elem()
	val := reflect.ValueOf(v).Elem()
	nums := val.NumField()
	structFields := make([]reflect.StructField, 0)
	fieldMap := make(map[string]reflect.Value)
	for i := 0; i < nums; i++ {
		if typ.Field(i).Name != targetName {
			structFields = append(structFields, reflect.StructField{
				Name: typ.Field(i).Name,
				Type: typ.Field(i).Type,
			})
			fieldMap[typ.Field(i).Name] = val.Field(i)
		}
	}
	newtyp := reflect.StructOf(structFields)
	newi := reflect.New(newtyp).Elem()
	newty := reflect.TypeOf(newi.Addr().Interface()).Elem()
	for i := 0; i < nums-1; i++ {
		newi.Field(i).Set(fieldMap[newty.Field(i).Name])
	}
	return newi.Addr().Interface()
}

func getJavaClassName(pkg interface{}) string {
	val := reflect.ValueOf(pkg).Elem()
	typ := reflect.TypeOf(pkg).Elem()
	nums := val.NumField()
	for i := 0; i < nums; i++ {
		if typ.Field(i).Name == "JavaClassName" {
			return val.Field(i).String()
		}
	}
	fmt.Println("error: JavaClassName not found")
	return ""
}
