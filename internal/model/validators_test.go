package model

import (
	"reflect"
	"testing"
)

func TestValidator_Kind(t *testing.T) {
	v := kind(reflect.String)

	ok, et := v("ok")
	if !ok {
		t.Errorf("Expected validation of \"ok\" to pass, but \"%s\"", *et)
	}

	ok, et = v(1)
	if ok {
		t.Errorf("Expected validation of 1 to fail, but it passed")
	}
	expectedTrace := "validation failed at:  (expected kind: `string`, got value `1` of kind `int`)"
	if et.String() != expectedTrace {
		t.Errorf("Expected error trace \"%s\", got \"%s\"", expectedTrace, *et)
	}

	ok, et = v(nil)
	if ok {
		t.Errorf("Expected validation of <nil> to fail, but it passed")
	}
	expectedTrace = "validation failed at:  (expected kind: `string`, got value `<nil>` of kind `unknown`)"
	if et.String() != expectedTrace {
		t.Errorf("Expected error trace \"%s\", got \"%s\"", expectedTrace, *et)
	}
}

func TestValidator_Typ(t *testing.T) {
	type MyType string
	v := typ(MyType(""))

	ok, et := v(MyType("another value"))
	if !ok {
		t.Errorf("Expected validation of MyType(\"another value\") to pass, but \"%s\"", *et)
	}

	ok, et = v("my str")
	if ok {
		t.Errorf("Expected validation of \"my str\" to fail, but it passed")
	}
	expectedTrace := "validation failed at:  (expected type: `model.MyType`, got value `my str` of type `string`)"
	if et.String() != expectedTrace {
		t.Errorf("Expected error trace \"%s\", got \"%s\"", expectedTrace, *et)
	}

	ok, et = v(nil)
	if ok {
		t.Errorf("Expected validation of <nil> to fail, but it passed")
	}
	expectedTrace = "validation failed at:  (expected type: `model.MyType`, got value `<nil>` of type `unknown`)"
	if et.String() != expectedTrace {
		t.Errorf("Expected error trace \"%s\", got \"%s\"", expectedTrace, *et)
	}
}
