package goconfig

import (
	"fmt"
	"reflect"
	"strconv"
)

type node struct {
	parent *node

	hasValue    bool
	tags        map[string][]string
	actualType  reflect.Type
	field       reflect.StructField
	value       reflect.Value
	actualValue reflect.Value
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

func flatten(root reflect.Value) []*node {
	result := []*node{}
	current := &node{
		actualType:  indirect(root.Type()),
		value:       root,
		actualValue: reflect.Indirect(root),
		field:       reflect.StructField{Name: "root", Type: root.Type()},
	}
	toVisit := []*node{current}
	seenTypes := map[string]bool{}
	for len(toVisit) > 0 {
		current, toVisit = toVisit[0], toVisit[1:]
		result = append(result, current)
		key := current.actualType.Name() + current.field.Name
		if seenTypes[key] || current.actualType.Kind() != reflect.Struct {
			continue
		}
		seenTypes[key] = true
		for _, f := range reflect.VisibleFields(current.actualType) {
			n := &node{
				field:      f,
				parent:     current,
				actualType: indirect(f.Type),
				tags:       collectTags(f, current.tags),
			}
			if current.actualValue.IsValid() && !current.actualValue.IsZero() {
				n.value = current.actualValue.FieldByName(f.Name)
				n.actualValue = reflect.Indirect(n.value)
			}
			toVisit = append(toVisit, n)
		}
	}
	return result
}

func getScalars(nodes []*node) []*node {
	result := []*node{}
	for _, node := range nodes {
		if node.actualType.Kind() == reflect.Struct {
			continue
		}
		result = append(result, node)
	}
	return result
}

func resolveParents(n *node) {
	for n.parent != nil {
		if !n.parent.value.IsValid() {
			n.parent.value = reflect.New(n.parent.actualType)
		}
		if !n.parent.actualValue.IsValid() {
			n.parent.actualValue = reflect.Indirect(n.parent.value)
		}
		n.parent.actualValue.
			FieldByName(n.field.Name).
			Set(n.value)
		n = n.parent
	}
}

func set(value *reflect.Value, str string) error {
	var result error
	kind := value.Kind()
	switch kind {
	case reflect.Bool:
		asBool, err := strconv.ParseBool(str)
		result = err
		value.SetBool(asBool)
	case reflect.String:
		value.SetString(str)
	case reflect.Int, reflect.Int64:
		asInt, err := strconv.Atoi(str)
		result = err
		value.SetInt(int64(asInt))
	case reflect.Uint, reflect.Uint64:
		asUint, err := strconv.ParseUint(str, 10, 64)
		result = err
		value.SetUint(asUint)
	case reflect.Float32, reflect.Float64:
		asFloat, err := strconv.ParseFloat(str, 64)
		result = err
		value.SetFloat(asFloat)
	default:
		result = fmt.Errorf("unsupported field type %s", kind)
	}
	return result
}
