package model

import (
	"reflect"
	"testing"
)

func assertSuccess(t *testing.T, valueStr string, ok bool, et *errortrace) {
	if !ok {
		t.Errorf("Expected validation of %s to pass, but \"%s\"", valueStr, *et)
	}
}

func assertFailure(t *testing.T, valueStr string, ok bool, et *errortrace, expectedTrace string) {
	if ok {
		t.Errorf("Expected validation of %s to fail, but it passed", valueStr)
	}
	if et.String() != expectedTrace {
		t.Errorf("Expected error trace \"%s\", got \"%s\"", expectedTrace, *et)
	}
}

func TestValidator_Kind(t *testing.T) {
	v := kind(reflect.String)

	ok, et := v("ok")
	assertSuccess(t, "\"ok\"", ok, et)

	ok, et = v(1)
	assertFailure(t, "1", ok, et,
		"validation failed at:  (expected kind `string`, got value `1` of kind `int`)")

	ok, et = v(nil)
	assertFailure(t, "<nil>", ok, et,
		"validation failed at:  (expected kind `string`, got value `<nil>` of kind `unknown`)")
}

func TestValidator_Typ(t *testing.T) {
	type MyType string
	v := typ(MyType(""))

	ok, et := v(MyType("another value"))
	assertSuccess(t, "MyType(\"another value\")", ok, et)

	ok, et = v("my str")
	assertFailure(t, "MyType(\"another value\")", ok, et,
		"validation failed at:  (expected type `model.MyType`, got value `my str` of type `string`)")

	ok, et = v(nil)
	assertFailure(t, "<nil>", ok, et,
		"validation failed at:  (expected type `model.MyType`, got value `<nil>` of type `unknown`)")
}

func TestValidator_Opt(t *testing.T) {
	v := opt(kind(reflect.String))

	ok, et := v(1, false)
	assertSuccess(t, "1", ok, et)

	ok, et = v("", true)
	assertSuccess(t, "\"\"", ok, et)

	ok, et = v("")
	assertSuccess(t, "\"\"", ok, et)

	ok, et = v(1, true)
	assertFailure(t, "1", ok, et,
		"validation failed at:  (expected kind `string`, got value `1` of kind `int`)")

	ok, et = v(1)
	assertFailure(t, "1", ok, et,
		"validation failed at:  (expected kind `string`, got value `1` of kind `int`)")
}

func TestValidator_List(t *testing.T) {
	v := list(kind(reflect.String))

	// NB: json.Unmarshal turns json arrays into []interface{}
	ok, et := v([]interface{}{})
	assertSuccess(t, "[]", ok, et)

	ok, et = v([]interface{}{"str1", "str2"})
	assertSuccess(t, "[\"str1\", \"str2\"]", ok, et)

	ok, et = v([]interface{}{"str1", 1})
	assertFailure(t, "[\"str1\", 1]", ok, et,
		"validation failed at: [1] (expected kind `string`, got value `1` of kind `int`)")
}

func TestValidator_Doc(t *testing.T) {
	v := doc(map[string]validator{
		"f1": kind(reflect.String),
		"f2": opt(kind(reflect.String)),
		"f3": list(kind(reflect.String)),
	}, "MyDocumentName")

	value := map[string]interface{}{
		"f1": 1,
		"f3": []interface{}{},
	}

	ok, et := v(value)
	assertFailure(t, "{f1: 1, f3: []}", ok, et,
		"validation failed at: MyDocumentName.f1 (expected kind `string`, got value `1` of kind `int`)")

	value = map[string]interface{}{
		"f1": "str1",
		"f2": int64(1),
		"f3": []interface{}{"str2"},
	}

	ok, et = v(value)
	assertFailure(t, "{f1: 1, f2: 1.0, f3: [\"str2\"]}", ok, et,
		"validation failed at: MyDocumentName.f2 (expected kind `string`, got value `1` of kind `int64`)")

	value = map[string]interface{}{
		"f1": "str1",
		"f2": "str2",
		"f3": []interface{}{"str3", 3, "str4"},
	}
	ok, et = v(value)
	assertFailure(t, "{f1: \"str1\", f2: \"str2\", f3: [\"str3\", 3, \"str4\"]}", ok, et,
		"validation failed at: MyDocumentName.f3[1] (expected kind `string`, got value `3` of kind `int`)")
}
