package model

import (
	"reflect"
	"testing"
)

type A struct {
	F1 string
}

type B struct {
	F1 int
}

type C struct {
	F1 A
	F2 map[string]int
	F3 []int
	F4 map[string]A
	F5 map[string][]int
	F6 map[string]map[string]int
	F7 []A
	F8 []map[string]int
	F9 [][]int
}

type D struct {
	F1 int `validation:"required"`
}

type E struct {
	F1 int `validation:"internal"`
}

func TestValiation_ValidJson(t *testing.T) {
	// invalid json
	raw := []byte(`{"Foo", 1 }`)
	et := Validate(reflect.TypeOf(A{}), raw)
	if et == nil || et.Error() != "validation failed: `{\"Foo\", 1 }` is invalid json" {
		t.Errorf("expected validation of %s to fail", raw)
	}
}

func TestValidation_Kind(t *testing.T) {
	// json.Number - got number, want something else
	raw := []byte(`{"F1": 1}`)
	et := Validate(reflect.TypeOf(A{}), raw)
	if et == nil || et.Error() != "validation failed at A.F1: expected kind string, got value `1` of type json.Number" {
		t.Errorf("expected validation of %s to fail", raw)
	}

	// json.Number - cannot convert number into kind
	raw = []byte(`{"F1": 1.0}`)
	et = Validate(reflect.TypeOf(B{}), raw)
	if et == nil || et.Error() != "validation failed at B.F1: cannot convert value `1.0` into kind int" {
		t.Errorf("expected validation of %s to fail", raw)
	}

	// wrong kind, other
	raw = []byte(`{"F1": false}`)
	et = Validate(reflect.TypeOf(A{}), raw)
	if et == nil || et.Error() != "validation failed at A.F1: expected kind string, got value `false` of kind bool" {
		t.Errorf("expected validation of %s to fail", raw)
	}

	// null value
	raw = []byte(`{"F1": null}`)
	et = Validate(reflect.TypeOf(A{}), raw)
	if et == nil || et.Error() != "validation failed at A.F1: unexpected null value" {
		t.Errorf("expected validation of %s to fail", raw)
	}
}

func TestValidation_Struct(t *testing.T) {
	// unexpected field
	raw := []byte(`{"F2": 1}`)
	et := Validate(reflect.TypeOf(A{}), raw)
	if et == nil || et.Error() != "validation failed at A: found unexpected field `F2`" {
		t.Errorf("expected validation of %s to fail", raw)
	}

	// required field
	raw = []byte(`{}`)
	et = Validate(reflect.TypeOf(D{}), raw)
	if et == nil || et.Error() != "validation failed at D: field F1 is required, but not specified" {
		t.Errorf("expected validation of %s to fail", raw)
	}

	// internal field
	raw = []byte(`{"F1": 1}`)
	et = Validate(reflect.TypeOf(E{}), raw)
	if et == nil || et.Error() != "validation failed at E: found unexpected field `F1`" {
		t.Errorf("expected validation of %s to fail", raw)
	}
}

func TestValidation_Nesting(t *testing.T) {
	// struct field
	raw := []byte(`{"F1": {"F1": 1}}`)
	et := Validate(reflect.TypeOf(C{}), raw)
	if et == nil || et.Error() != "validation failed at C.F1.F1: expected kind string, got value `1` of type json.Number" {
		t.Errorf("expected validation of %s to fail", raw)
	}

	// map field
	raw = []byte(`{"F2": {"K1": "hello"}}`)
	et = Validate(reflect.TypeOf(C{}), raw)
	if et == nil || et.Error() != "validation failed at C.F2[\"K1\"]: expected kind int, got value `hello` of kind string" {
		t.Errorf("expected validation of %s to fail", raw)
	}

	// slice field
	raw = []byte(`{"F3": ["hello"]}`)
	et = Validate(reflect.TypeOf(C{}), raw)
	if et == nil || et.Error() != "validation failed at C.F3[0]: expected kind int, got value `hello` of kind string" {
		t.Errorf("expected validation of %s to fail", raw)
	}

	// map of struct
	raw = []byte(`{"F4": {"K1": {"F1": 1}}}`)
	et = Validate(reflect.TypeOf(C{}), raw)
	if et == nil || et.Error() != "validation failed at C.F4[\"K1\"].F1: expected kind string, got value `1` of type json.Number" {
		t.Errorf("expected validation of %s to fail", raw)
	}

	// map of slice
	raw = []byte(`{"F5": {"K1": ["hello"]}}`)
	et = Validate(reflect.TypeOf(C{}), raw)
	if et == nil || et.Error() != "validation failed at C.F5[\"K1\"][0]: expected kind int, got value `hello` of kind string" {
		t.Errorf("expected validation of %s to fail", raw)
	}

	// map of map
	raw = []byte(`{"F6": {"K1": {"K1": "hello"}}}`)
	et = Validate(reflect.TypeOf(C{}), raw)
	if et == nil || et.Error() != "validation failed at C.F6[\"K1\"][\"K1\"]: expected kind int, got value `hello` of kind string" {
		t.Errorf("expected validation of %s to fail", raw)
	}

	// slice of struct
	raw = []byte(`{"F7": [{"F1": 1}]}`)
	et = Validate(reflect.TypeOf(C{}), raw)
	if et == nil || et.Error() != "validation failed at C.F7[0].F1: expected kind string, got value `1` of type json.Number" {
		t.Errorf("expected validation of %s to fail", raw)
	}

	// slice of map
	raw = []byte(`{"F8": [{"K1": "hello"}]}`)
	et = Validate(reflect.TypeOf(C{}), raw)
	if et == nil || et.Error() != "validation failed at C.F8[0][\"K1\"]: expected kind int, got value `hello` of kind string" {
		t.Errorf("expected validation of %s to fail", raw)
	}

	// slice of slice
	raw = []byte(`{"F9": [[1, 2], [3, 4, 5, "hi"]]}`)
	et = Validate(reflect.TypeOf(C{}), raw)
	if et == nil || et.Error() != "validation failed at C.F9[1][3]: expected kind int, got value `hi` of kind string" {
		t.Errorf("expected validation of %s to fail", raw)
	}
}
