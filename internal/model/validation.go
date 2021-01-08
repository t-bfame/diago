package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// ErrorTrace stores info about what went wrong during
// the validation check
type ErrorTrace struct {
	reason  string
	context []string
}

func (et *ErrorTrace) Error() string {
	var sb strings.Builder
	if len(et.context) > 0 {
		sb.WriteString("validation failed at ")
		for i := len(et.context) - 1; i >= 0; i-- {
			sb.WriteString(et.context[i])
		}
		sb.WriteString(": ")
	} else {
		sb.WriteString("validation failed: ")
	}
	sb.WriteString(et.reason)
	return sb.String()
}

func (et *ErrorTrace) attach(msg string) {
	et.context = append(et.context, msg)
}

// Validate checks if blob can be unmarshalled into valid struct of type t
func Validate(t reflect.Type, blob []byte) error {
	// confirm is valid json
	var v interface{}
	decoder := json.NewDecoder(bytes.NewBuffer(blob))
	decoder.UseNumber()
	if err := decoder.Decode(&v); err != nil {
		return &ErrorTrace{
			reason: fmt.Sprintf("`%s` is invalid json", blob),
		}
	}

	trace := &ErrorTrace{}
	if !structure(t, v, trace) {
		trace.attach(t.Name())
		return trace
	}

	return nil
}

func structure(t reflect.Type, v interface{}, trace *ErrorTrace) bool {
	// check is map
	if !kind(reflect.Map, v, trace) {
		return false
	}

	// any unexpected fields?
	for k := range v.(map[string]interface{}) {
		f, exists := t.FieldByName(k)
		internal := false
		if exists {
			// disable specification of internal fields
			tags, ok := f.Tag.Lookup("validation")
			if ok {
				for _, t := range strings.Split(tags, ",") {
					if t == "internal" {
						internal = true
					}
				}
			}
		}
		if !exists || internal {
			trace.reason = fmt.Sprintf("found unexpected field `%s`", k)
			return false
		}
	}

	// all fields are valid?
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fv, exists := v.(map[string]interface{})[f.Name]
		if !exists {
			required := false
			tags, ok := f.Tag.Lookup("validation")
			if ok {
				for _, t := range strings.Split(tags, ",") {
					if t == "required" {
						required = true
					}
				}
			}
			if required {
				trace.reason = fmt.Sprintf("field %s is required, but not specified", f.Name)
				return false
			}
			continue
		}
		switch f.Type.Kind() {
		case reflect.Map:
			if !mapping(f.Type.Elem(), fv, trace) {
				trace.attach(fmt.Sprintf(".%s", f.Name))
				return false
			}
		case reflect.Struct:
			if !structure(f.Type, fv, trace) {
				trace.attach(fmt.Sprintf(".%s", f.Name))
				return false
			}
		case reflect.Slice:
			if !slice(f.Type.Elem(), fv, trace) {
				trace.attach(fmt.Sprintf(".%s", f.Name))
				return false
			}
		default:
			if !kind(f.Type.Kind(), fv, trace) {
				trace.attach(fmt.Sprintf(".%s", f.Name))
				return false
			}
		}
	}

	return true
}

func kind(k reflect.Kind, v interface{}, trace *ErrorTrace) bool {
	if v == nil {
		trace.reason = "unexpected null value"
		return false
	}

	vK := reflect.TypeOf(v).Kind()

	// NB: By default, json unmarshals into float64, so
	// we use a decoder to decode into json.Number.
	// We need to check for convertibility instead of
	// equality since json.Number's underlying kind is String.
	if reflect.TypeOf(v) == reflect.TypeOf(json.Number("")) {
		var err error
		switch k {
		case reflect.Int:
			_, err = strconv.Atoi(string(v.(json.Number)))
		case reflect.Int8:
			_, err = strconv.ParseInt(string(v.(json.Number)), 10, 8)
		case reflect.Int16:
			_, err = strconv.ParseInt(string(v.(json.Number)), 10, 16)
		case reflect.Int32:
			_, err = strconv.ParseInt(string(v.(json.Number)), 10, 32)
		case reflect.Int64:
			_, err = strconv.ParseInt(string(v.(json.Number)), 10, 64)
		case reflect.Uint:
			// The size of uint depends on implementation.
			// Let's just make sure it fits into 32 bits for now.
			_, err = strconv.ParseUint(string(v.(json.Number)), 10, 32)
		case reflect.Uint8:
			_, err = strconv.ParseUint(string(v.(json.Number)), 10, 8)
		case reflect.Uint16:
			_, err = strconv.ParseUint(string(v.(json.Number)), 10, 16)
		case reflect.Uint32:
			_, err = strconv.ParseUint(string(v.(json.Number)), 10, 32)
		case reflect.Uint64:
			_, err = strconv.ParseUint(string(v.(json.Number)), 10, 64)
		case reflect.Float32:
			_, err = strconv.ParseFloat(string(v.(json.Number)), 32)
		case reflect.Float64:
			_, err = strconv.ParseFloat(string(v.(json.Number)), 64)
		default:
			trace.reason = fmt.Sprintf("expected kind %s, got value `%v` of type json.Number", k, v)
			return false
		}

		if err != nil {
			trace.reason = fmt.Sprintf("cannot convert value `%v` into kind %s", v, k)
			return false
		}

		return true
	}

	if k != vK {
		trace.reason = fmt.Sprintf("expected kind %s, got value `%v` of kind %s", k, v, vK)
		return false
	}

	return true
}

func mapping(elemT reflect.Type, v interface{}, trace *ErrorTrace) bool {
	// check is map
	if !kind(reflect.Map, v, trace) {
		return false
	}

	for ek, ev := range v.(map[string]interface{}) {
		switch elemT.Kind() {
		case reflect.Map:
			if !mapping(elemT.Elem(), ev, trace) {
				trace.attach(fmt.Sprintf("[\"%s\"]", ek))
				return false
			}
		case reflect.Struct:
			if !structure(elemT, ev, trace) {
				trace.attach(fmt.Sprintf("[\"%s\"]", ek))
				return false
			}
		case reflect.Slice:
			if !slice(elemT.Elem(), ev, trace) {
				trace.attach(fmt.Sprintf("[\"%s\"]", ek))
				return false
			}
		default:
			if !kind(elemT.Kind(), ev, trace) {
				trace.attach(fmt.Sprintf("[\"%s\"]", ek))
				return false
			}
		}
	}

	return true
}

func slice(elemT reflect.Type, v interface{}, trace *ErrorTrace) bool {
	// check if slice
	if !kind(reflect.Slice, v, trace) {
		return false
	}

	for i, ev := range v.([]interface{}) {
		switch elemT.Kind() {
		case reflect.Map:
			if !mapping(elemT.Elem(), ev, trace) {
				trace.attach(fmt.Sprintf("[%d]", i))
				return false
			}
		case reflect.Struct:
			if !structure(elemT, ev, trace) {
				trace.attach(fmt.Sprintf("[%d]", i))
				return false
			}
		case reflect.Slice:
			if !slice(elemT.Elem(), ev, trace) {
				trace.attach(fmt.Sprintf("[%d]", i))
				return false
			}
		default:
			if !kind(elemT.Kind(), ev, trace) {
				trace.attach(fmt.Sprintf("[%d]", i))
				return false
			}
		}
	}

	return true
}
