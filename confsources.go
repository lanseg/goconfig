package goconfig

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

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
		}
	}
	return errors.Join(result...)
}

func FromArgs(nodes []*node) error {
	fargs := flag.NewFlagSet("Command line arguments", flag.ContinueOnError)
	leaves := map[string]*reflect.Value{}
	for _, node := range nodes {
		if node.field.Type.Kind() == reflect.Struct {
			continue
		}
		name := strings.Join(node.tags["arg"], "_")
		leaves[name] = &node.value
		fargs.String(name, "", "help")
	}
	if err := fargs.Parse(os.Args[1:]); err != nil {
		return err
	}
	result := []error{}
	fargs.Visit(func(f *flag.Flag) {
		value, ok := leaves[f.Name]
		if !ok {
			return
		}
		v := *value
		if v.Kind() == reflect.Pointer {
			v = v.Elem()
		}
		if err := set(&v, f.Value.String()); err != nil {
			result = append(result, err)
		}
	})
	return errors.Join(result...)
}
