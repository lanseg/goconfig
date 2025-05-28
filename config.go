package goconfig

import (
	"fmt"
	"reflect"
)

const (
	ParamSeparator = "_"
)

var (
	supportedTags = []string{"env", "arg"}
)

type ConfigSource = func(nodes []*node) error

func GetConfig[T any](sources ...ConfigSource) (*T, error) {
	return GetConfigTo(new(T), sources...)
}

func GetConfigTo[T any](root *T, sources ...ConfigSource) (*T, error) {
	rootValue := reflect.ValueOf(root)
	if reflect.Indirect(rootValue).Kind() != reflect.Struct {
		return nil, fmt.Errorf("only struct types are supported, but got kind %s", rootValue.Kind())
	}
	if len(sources) == 0 {
		return root, nil
	}
	nodes := flatten(rootValue)

	scalars := getScalars(nodes)
	for _, node := range scalars {
		node.value = reflect.New(indirect(node.field.Type)).Elem()
		node.actualValue = reflect.Indirect(node.value)
	}

	for _, src := range sources {
		if err := src(scalars); err != nil {
			return nil, err
		}
	}
	for _, node := range scalars {
		if !node.hasValue {
			continue
		}
		resolveParents(node)
		node.parent.actualValue.FieldByName(node.field.Name).Set(node.actualValue)
	}

	return root, nil
}
