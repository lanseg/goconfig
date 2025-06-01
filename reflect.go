package goconfig

import (
	"fmt"
	"reflect"
	"strconv"
)

var (
	primitiveKinds = map[reflect.Kind]bool{
		reflect.Bool: true, reflect.Int: true, reflect.Int8: true, reflect.Int16: true,
		reflect.Int32: true, reflect.Int64: true, reflect.Uint: true, reflect.Uint8: true,
		reflect.Uint16: true, reflect.Uint32: true, reflect.Uint64: true, reflect.Float32: true,
		reflect.Float64: true, reflect.String: true,
		// TODO: Support for Complex64, Complex128
	}
	withElem = map[reflect.Kind]bool{
		reflect.Array: true, reflect.Chan: true, reflect.Map: true, reflect.Pointer: true,
		reflect.Slice: true,
	}
	supportedTags = []string{"env", "arg"}
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
	for withElem[root.Kind()] {
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
		field:       reflect.StructField{Name: "root", Type: root.Type()},
		actualType:  indirect(root.Type()),
		value:       root,
		actualValue: reflect.Indirect(root),
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
				actualType: indirect(f.Type),
				parent:     current,
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
		if primitiveKinds[node.actualType.Kind()] {
			result = append(result, node)
		}
	}
	return result
}

func resolveParents(nodes []*node) {
	for _, node := range nodes {
		if !node.hasValue {
			continue
		}
		for n := node; n.parent != nil; n = n.parent {
			p := n.parent
			if !p.value.IsValid() {
				p.value = reflect.New(p.actualType)
				p.actualValue = reflect.Indirect(p.value)
			}
			p.actualValue.FieldByName(n.field.Name).Set(n.value)
		}
		node.parent.actualValue.FieldByName(node.field.Name).Set(node.actualValue)
	}
}

func hasCycles(nodes []*node) bool {
	// TODO: Implement properly
	for _, node := range nodes {
		for n := node.parent; n != nil; n = n.parent {
			if node.actualType == n.actualType {
				return true
			}
		}
	}
	return false
}

func set(value *reflect.Value, str string) error {
	var result error
	switch value.Kind() {
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
		result = fmt.Errorf("unsupported field type %s", value.Kind())
	}
	return result
}
