package goconfig

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"
)

func FromEnv(nodes []*node) error {
	result := []error{}
	for _, node := range nodes {
		if node.field.Type.Kind() == reflect.Struct {
			continue
		}
		varValue, ok := os.LookupEnv(strings.Join(node.tags["env"], "_"))
		if !ok {
			continue
		}
		v := node.value
		if v.Kind() == reflect.Pointer {
			v = v.Elem()
		}
		if err := set(&v, varValue); err != nil {
			result = append(result, err)
		} else {
			node.hasValue = true
		}
	}
	return errors.Join(result...)
}

func FromFlags(nodes []*node) error {
	return (&FlagSource{}).Collect(nodes)
}

type FlagSource struct {
	flags *flag.FlagSet
}

func (ff *FlagSource) Args() []string {
	if ff.flags == nil {
		return []string{}
	}
	return ff.flags.Args()
}

func setFunc(n *node) func(value string) error {
	return func(value string) error {
		if value == "" {
			return nil
		}
		if err := set(&n.value, value); err != nil {
			return err
		}
		n.hasValue = true
		return nil
	}
}

func (ff *FlagSource) Collect(nodes []*node) error {
	ff.flags = flag.NewFlagSet("Command line arguments", flag.ContinueOnError)
	for _, node := range nodes {
		nodeType := node.actualType
		if nodeType.Kind() == reflect.Struct {
			continue
		}
		ff.flags.BoolFunc(
			strings.Join(node.tags["arg"], "_"),
			fmt.Sprintf("parameter of type %q", node.actualType.Name()),
			setFunc(node))
	}
	return ff.flags.Parse(os.Args[1:])
}
