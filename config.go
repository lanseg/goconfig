package goconfig

import (
	"errors"
	"fmt"
	"reflect"
)

func GetConfig[T any](sources ...ConfigSource) (*T, error) {
	return GetConfigTo(new(T), sources...)
}

func GetConfigTo[T any](root *T, sources ...ConfigSource) (*T, error) {
	if len(sources) == 0 {
		return root, nil
	}
	if root == nil {
		return nil, errors.New("cannot use nil for the default struct")
	}
	rootValue := reflect.Indirect(reflect.ValueOf(root))
	if rootValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("only struct types are supported, but got kind %s", rootValue.Kind())
	}
	nodes := flatten(rootValue)
	if hasCycles(nodes) {
		return nil, errors.New("cycles found")
	}

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
	resolveParents(scalars)
	return root, nil
}
