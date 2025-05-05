package goconfig

import (
	"fmt"
	"reflect"
)

const (
	ParamSeparator = "_"
)

type ConfigSource = func(values []*valueDef) error

type pair[T any] struct {
	a T
	b T
}

func resolveTypeStruct(root reflect.Type) error {
	toVisit := []reflect.Type{root}
	nodes := map[reflect.Type]bool{}
	for len(toVisit) > 0 {
		current := toVisit[0]
		toVisit = toVisit[1:]
		for current.Kind() == reflect.Pointer {
			current = current.Elem()
		}
		if current.Kind() != reflect.Struct {
			continue
		}
		nodes[current] = true
		for _, field := range reflect.VisibleFields(current) {
			fieldType := field.Type
			for fieldType.Kind() == reflect.Pointer {
				fieldType = fieldType.Elem()
			}
			edge := pair[reflect.Type]{a: current, b: fieldType}
			if !edges[edge] {
				edges[edge] = true
				toVisit = append(toVisit, field.Type)
			}
		}
	}
	for n := range nodes {
		fmt.Printf("HERE-Node %v\n", n)
	}
	for e := range edges {
		fmt.Printf("HERE-Edge %v -> %v\n", e.a, e.b)
	}
	fmt.Println()
	return nil
}

type valueDef struct {
	field *reflect.StructField
	value *reflect.Value
	tags  map[string][]string
	leaf  bool
}

func newValueDef(parent reflect.Value, field int, tags map[string][]string) *valueDef {
	v, f := parent.Field(field), parent.Type().Field(field)
	result := &valueDef{value: &v, field: &f, tags: map[string][]string{}}
	for _, tag := range []string{"env", "arg"} {
		result.tags[tag] = append(tags[tag], f.Tag.Get(tag))
	}
	return result
}

func resolveTypeFields[T any]() (*T, []*valueDef) {
	obj := new(T)
	root := reflect.ValueOf(obj)
	if err := resolveTypeStruct(root.Type()); err != nil {
		fmt.Println(err)
	}
	result := []*valueDef{}
	toVisit := []*valueDef{{value: &root, field: nil, tags: map[string][]string{}}}
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
				toVisit = append(toVisit, newValueDef(actualValue, fi, current.tags))
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
