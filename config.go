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

type Set[T comparable] map[T]bool

func (s Set[T]) Add(value T) {
	s[value] = true
}

func (s Set[T]) Has(value T) bool {
	return s[value]
}

type node struct {
	parent *node

	tags  map[string][]string
	field reflect.StructField
	value reflect.Value
}

func indirect(root reflect.Type) reflect.Type {
	for root.Kind() == reflect.Pointer || root.Kind() == reflect.Slice || root.Kind() == reflect.Array {
		root = root.Elem()
	}
	return root
}

func collectTags(field reflect.StructField, base map[string][]string) map[string][]string {
	result := map[string][]string{}
	for _, tagName := range supportedTags {
		if tagValue, ok := field.Tag.Lookup(tagName); ok {
			result[tagName] = append(base[tagName], tagValue)
		} else {
			result[tagName] = append(base[tagName], field.Name)
		}
	}
	return result
}

func flatten(root reflect.Type) []*node {
	result := []*node{}
	current := &node{field: reflect.StructField{Name: "root", Type: root}}
	toVisit := []*node{current}
	seenTypes := Set[string]{}
	for len(toVisit) > 0 {
		current, toVisit = toVisit[0], toVisit[1:]
		result = append(result, current)
		actualType := indirect(current.field.Type)
		key := actualType.Name() + current.field.Name
		if seenTypes.Has(key) || actualType.Kind() != reflect.Struct {
			continue
		}
		seenTypes.Add(key)
		for _, f := range reflect.VisibleFields(actualType) {
			n := &node{
				field:  f,
				parent: current,
				tags:   collectTags(f, current.tags),
			}
			toVisit = append(toVisit, n)
		}
	}
	return result
}

type ConfigSource = func(nodes []*node) error

func GetConfig[T any](sources ...ConfigSource) (*T, error) {
	result := new(T)
	nodes := flatten(reflect.TypeFor[T]())
	nodes[0].value = reflect.ValueOf(result)
	for i, node := range nodes[1:] {
		nodes[i+1].value = reflect.New(indirect(node.field.Type))
	}
	for _, src := range sources {
		if err := src(nodes); err != nil {
			return nil, err
		}
	}
	for _, node := range nodes[1:] {
		parent := node.parent.value
		if parent.Kind() == reflect.Pointer {
			parent = parent.Elem()
		}
		childField := parent.FieldByName(node.field.Name)
		if childField.Kind() == reflect.Pointer && node.value.Kind() != reflect.Pointer {
			childField.Elem().Set(node.value)
		} else if childField.Kind() != reflect.Pointer && node.value.Kind() == reflect.Pointer {
			childField.Set(node.value.Elem())
		} else {
			childField.Set(node.value)
		}
	}
	return result, nil
}
