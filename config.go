package goconfig

import (
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
	result := new(T)
	nodes := flatten(reflect.TypeFor[T]())
	nodes[0].value = reflect.ValueOf(result)

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

	return result, nil
}
