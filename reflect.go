package goconfig

import "reflect"

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

func flatten(root reflect.Type) []*node {
	result := []*node{}
	current := &node{
		actualType: indirect(root),
		field:      reflect.StructField{Name: "root", Type: root},
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
