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

func FromEnv(values []*valueDef) error {
	result := []error{}
	for _, value := range values {
		if !value.leaf {
			continue
		}
		varValue, ok := os.LookupEnv(strings.Join(value.tags["env"], "_"))
		if !ok {
			continue
		}
		if err := set(value.value, varValue); err != nil {
			result = append(result, err)
		}
	}
	return errors.Join(result...)
}

func FromArgs(values []*valueDef) error {
	fargs := flag.NewFlagSet("Command line arguments", flag.ContinueOnError)
	leaves := map[string]*valueDef{}
	for _, value := range values {
		if !value.leaf {
			continue
		}
		name := strings.Join(value.tags["arg"], "_")
		leaves[name] = value
		fargs.String(name, "", "help")
	}
	if err := fargs.Parse(os.Args[1:]); err != nil {
		return err
	}
	result := []error{}
	fargs.Visit(func(f *flag.Flag) {
		leaf, ok := leaves[f.Name]
		if !ok {
			return
		}
		if err := set(leaf.value, f.Value.String()); err != nil {
			result = append(result, err)
		}
	})
	return errors.Join(result...)
}
