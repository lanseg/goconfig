package goconfig

import (
	"os"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type Scalars struct {
	Bool    bool    `arg:"bool_field" env:"BOOL_FIELD"`
	String  string  `arg:"string_field" env:"STRING_FIELD"`
	Int     int     `arg:"int_field" env:"INT_FIELD"`
	Int64   int64   `arg:"int64_field" env:"INT64_FIELD"`
	Uint    uint    `arg:"uint_field" env:"UINT_FIELD"`
	Uint64  uint64  `arg:"uint64_field" env:"UINT64_FIELD"`
	Float32 float32 `arg:"float32_field" env:"FLOAT32_FIELD"`
	Float64 float64 `arg:"float64_field" env:"FLOAT64_FIELD"`
}

func TestPlainSettings(t *testing.T) {

	fullScalarArgs := []string{
		"--bool_field=true",
		"--string_field=String_field_set",
		"--int_field=-123",
		"--int64_field=-123456789",
		"--uint_field=123456789",
		"--uint64_field=123456789123456789",
		"--float32_field=3.1415",
		"--float64_field=3.141592653589793",
	}

	fullScalarArgResult := &Scalars{
		Bool:    true,
		String:  "String_field_set",
		Int:     -123,
		Int64:   -123456789,
		Uint:    123456789,
		Uint64:  123456789123456789,
		Float32: 3.1415,
		Float64: 3.141592653589793,
	}

	fullScalarEnv := map[string]string{
		"BOOL_FIELD":    "true",
		"STRING_FIELD":  "String_field_set",
		"INT_FIELD":     "-123",
		"INT64_FIELD":   "-123456789",
		"UINT_FIELD":    "123456789",
		"UINT64_FIELD":  "123456789123456789",
		"FLOAT32_FIELD": "3.1415",
		"FLOAT64_FIELD": "3.141592653589793",
	}

	fullScalarEnvResult := &Scalars{
		Bool:    true,
		String:  "String_field_set",
		Int:     -123,
		Int64:   -123456789,
		Uint:    123456789,
		Uint64:  123456789123456789,
		Float32: 3.1415,
		Float64: 3.141592653589793,
	}

	for _, tc := range []struct {
		name    string
		args    []string
		envs    map[string]string
		src     []ConfigSource
		want    *Scalars
		wantErr bool
	}{
		{
			name: "arg scalars all set",
			args: fullScalarArgs,
			want: fullScalarArgResult,
		},
		{
			name: "arg some scalars set",
			args: []string{"--string_field=String_field_set"},
			want: &Scalars{String: "String_field_set"},
		},
		{
			name: "env scalars all set",
			envs: fullScalarEnv,
			want: fullScalarEnvResult,
		},
		{
			name: "env some scalars set",
			envs: map[string]string{
				"BOOL_FIELD":    "true",
				"STRING_FIELD":  "Another string",
				"FLOAT64_FIELD": "3.141592653589793",
			},
			want: &Scalars{String: "Another string", Bool: true, Float64: 3.141592653589793},
		},
		{
			name: "scalars latest overrides env over args",
			args: fullScalarArgs,
			envs: fullScalarEnv,
			want: fullScalarEnvResult,
		},
		{
			name: "scalars latest overrides arg over env",
			args: fullScalarArgs,
			envs: fullScalarEnv,
			src:  []ConfigSource{FromEnv, FromArgs},
			want: fullScalarArgResult,
		},
		{
			name: "scalars empty result",
			args: []string{},
			envs: map[string]string{},
			want: &Scalars{},
		},
		{
			name:    "scalars wrong arg type should fail",
			args:    []string{"--bool_field=not_a_bool"},
			wantErr: true,
		},
		{
			name:    "scalars wrong env type should fail",
			envs:    map[string]string{"BOOL_FIELD": "not a bool"},
			wantErr: true,
		},
		{
			name:    "scalars wrong overridden by correct still fail arg first",
			args:    []string{"--bool_field=not_a_bool"},
			envs:    map[string]string{"BOOL_FIELD": "false"},
			src:     []ConfigSource{FromArgs, FromEnv},
			wantErr: true,
		},
		{
			name:    "scalars wrong overridden by correct still fail env first",
			args:    []string{"--bool_field=false"},
			envs:    map[string]string{"BOOL_FIELD": "not a bool"},
			src:     []ConfigSource{FromEnv, FromArgs},
			wantErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envs {
				t.Setenv(k, v)
			}
			os.Args = append([]string{os.Args[0]}, tc.args...)
			if tc.src == nil {
				tc.src = []ConfigSource{FromArgs, FromEnv}
			}
			config, err := GetConfig[Scalars](tc.src...)
			if err != nil && !tc.wantErr {
				t.Errorf("Unexpected error: %s", err)
			} else if !reflect.DeepEqual(tc.want, config) {
				t.Errorf("Expected config for flags %q should be \n%v, but got\n%v",
					tc.args, tc.want, config)
			}
		})
	}
}

type NestedScalarsInner struct {
	First  *Scalars `arg:"first" env:"FIRST"`
	Second *Scalars `arg:"second" env:"SECOND"`
	String string   `arg:"string" env:"STRING"`
}

type NestedScalarsOuter struct {
	First  *NestedScalarsInner `arg:"first" env:"FIRST"`
	Second *NestedScalarsInner `arg:"second" env:"SECOND"`
	Third  *Scalars            `arg:"third" env:"THIRD"`
	String string              `arg:"string" env:"STRING"`
}

func TestNestedSettings(t *testing.T) {
	for _, tc := range []struct {
		name    string
		args    []string
		envs    map[string]string
		src     []ConfigSource
		want    *NestedScalarsOuter
		wantErr bool
	}{
		{
			name: "arg some scalars set, others empty",
			args: []string{"--first_first_string_field=String_field_set"},
			want: &NestedScalarsOuter{
				First: &NestedScalarsInner{
					First:  &Scalars{String: "String_field_set"},
					Second: &Scalars{},
				},
				Second: &NestedScalarsInner{First: &Scalars{}, Second: &Scalars{}},
				Third:  &Scalars{},
			},
		},
		{
			name:    "arg name mismatch",
			args:    []string{"--first_string_field=String_field_set"},
			wantErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.envs {
				t.Setenv(k, v)
			}
			os.Args = append([]string{os.Args[0]}, tc.args...)
			if tc.src == nil {
				tc.src = []ConfigSource{FromArgs, FromEnv}
			}
			got, err := GetConfig[NestedScalarsOuter](tc.src...)
			if err != nil && !tc.wantErr {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Config mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

type RecursiveSettings struct {
	Value string
	Left  *RecursiveSettings
	Right *RecursiveSettings
}

func TestRecursiveSettings(t *testing.T) {
	t.Run("self recursed binary tree", func(t *testing.T) {
		_, err := GetConfig[RecursiveSettings](FromArgs, FromEnv)
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
			return
		}
	})
}
