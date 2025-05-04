package goconfig

import (
	"reflect"
)

const (
	ParamSeparator = "_"
)

type ConfigSource = func(values []*valueDef) error

type valueDef struct {
	field *reflect.StructField
	value *reflect.Value
	leaf  bool
	tags  map[string][]string
}

func newValueDef(v *reflect.Value, f *reflect.StructField, tags map[string][]string) *valueDef {
	result := &valueDef{value: v, field: f, tags: map[string][]string{}}
	if f == nil {
		return result
	}
	for _, tag := range []string{"env", "arg"} {
		result.tags[tag] = append(tags[tag], f.Tag.Get(tag))
	}
	return result
}

func resolveTypeFields[T any]() (*T, []*valueDef) {
	obj := new(T)
	root := reflect.ValueOf(obj)
	toVisit := []*valueDef{newValueDef(&root, nil, map[string][]string{})}
	result := []*valueDef{}
	for len(toVisit) > 0 {
		current := toVisit[0]
		result = append(result, current)
		if current.value.Kind() == reflect.Pointer && current.value.IsNil() {
			current.value.Set(reflect.New(current.value.Type().Elem()))
		}
		actualValue := *current.value
		if actualValue.Kind() == reflect.Pointer {
			actualValue = actualValue.Elem()
		}
		toVisit = toVisit[1:]
		if actualValue.Kind() == reflect.Struct {
			for fi := range actualValue.NumField() {
				a, b := actualValue.Field(fi), actualValue.Type().Field(fi)
				toVisit = append(toVisit, newValueDef(&a, &b, current.tags))
			}
		} else {
			current.leaf = true
		}
	}
	return obj, result
}

func GetConfig[T any](src ...ConfigSource) (*T, error) {
	result, fields := resolveTypeFields[T]()
	for _, updater := range src {
		if err := updater(fields); err != nil {
			return nil, err
		}
	}
	return result, nil
}
