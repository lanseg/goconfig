package goconfig

import (
	"errors"
	"flag"
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

type FlagSource struct {
	flags *flag.FlagSet
}

func (ff *FlagSource) Args() []string {
	if ff.flags == nil {
		return []string{}
	}
	return ff.flags.Args()
}

func (ff *FlagSource) Collect(nodes []*node) error {
	ff.flags = flag.NewFlagSet("Command line arguments", flag.ContinueOnError)
	leaves := map[string]*node{}
	for _, node := range nodes {
		if node.field.Type.Kind() == reflect.Struct {
			continue
		}
		name := strings.Join(node.tags["arg"], "_")
		leaves[name] = node
		ff.flags.String(name, "", "help")
	}
	if err := ff.flags.Parse(os.Args[1:]); err != nil {
		return err
	}
	result := []error{}
	ff.flags.Visit(func(f *flag.Flag) {
		n, ok := leaves[f.Name]
		if !ok {
			return
		}
		v := n.value
		if v.Kind() == reflect.Pointer {
			v = v.Elem()
		}
		if err := set(&v, f.Value.String()); err != nil {
			result = append(result, err)
		} else {
			n.hasValue = true
		}
	})
	return errors.Join(result...)
}
