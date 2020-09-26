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
		return
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

	ok, et = v(nil, false)
	assertFailure(t, "<nil>", ok, et,
		"validation failed at:  (field is required, but was not found)")
}

func TestValidator_Typ(t *testing.T) {
	type MyType string
	v := typ(MyType(""))

	ok, et := v(MyType("another value"))
	assertSuccess(t, "MyType(\"another value\")", ok, et)

	ok, et = v("my str")
	assertFailure(t, "\"my str\"", ok, et,
		"validation failed at:  (expected type `model.MyType`, got value `my str` of type `string`)")

	ok, et = v(nil)
	assertFailure(t, "<nil>", ok, et,
		"validation failed at:  (expected type `model.MyType`, got value `<nil>` of type `unknown`)")

	ok, et = v(nil, false)
	assertFailure(t, "<nil>", ok, et,
		"validation failed at:  (field is required, but was not found)")
}

func TestValidator_Opt(t *testing.T) {
	v := opt(kind(reflect.String))

	ok, et := v(nil, false)
	assertSuccess(t, "<nil>", ok, et)

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

	ok, et = v([]interface{}{"str1", 5})
	assertFailure(t, "[\"str1\", 5]", ok, et,
		"validation failed at: [1] (expected kind `string`, got value `5` of kind `int`)")
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
	assertFailure(t, "{f1: 1, f2: int64(1), f3: [\"str2\"]}", ok, et,
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

func TestValidator_Nested(t *testing.T) {
	v := doc(map[string]validator{
		"f1": opt(list(
			doc(map[string]validator{
				"i1": kind(reflect.Int),
				"i2": opt(list(kind(reflect.Int))),
			}),
		)),
	}, "MyDocumentName")

	value := map[string]interface{}{}
	ok, et := v(value)
	assertSuccess(t, "{}", ok, et)

	value = map[string]interface{}{
		"f1": 1,
	}
	ok, et = v(value)
	assertFailure(t, "{f1: 1}", ok, et,
		"validation failed at: MyDocumentName.f1 (expected kind `slice`, got value `1` of kind `int`)")

	value = map[string]interface{}{
		"f1": []interface{}{
			map[string]interface{}{},
		},
	}
	ok, et = v(value)
	assertFailure(t, "{f1: [{}]}", ok, et,
		"validation failed at: MyDocumentName.f1[0].i1 (field is required, but was not found)")

	value = map[string]interface{}{
		"f1": []interface{}{
			map[string]interface{}{
				"i1": 1,
			},
			map[string]interface{}{
				"i1": 1,
				"i2": []interface{}{"str1"},
			},
		},
	}
	ok, et = v(value)
	assertFailure(t, "{f1: [{i1: 1}, {i1: 1, i2: [\"str1\"]]}", ok, et,
		"validation failed at: MyDocumentName.f1[1].i2[0] (expected kind `int`, got value `str1` of kind `string`)")

	value = map[string]interface{}{
		"f1": []interface{}{
			map[string]interface{}{
				"i1": 1,
			},
			map[string]interface{}{
				"i1": 1,
				"i2": []interface{}{1, 2, 3},
			},
		},
	}
	ok, et = v(value)
	assertSuccess(t, "{f1: [{i1: 1}, {i1: 1, i2: [1, 2, 3]]}", ok, et)
}
