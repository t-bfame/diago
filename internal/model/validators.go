package model

import (
	"fmt"
	"reflect"
	"strings"
)

type validator func (value interface{}, exists ...bool) (bool, *errortrace)

type errortrace struct {
	trace	[]string
}

func (v errortrace) String() string {
	var sb strings.Builder
	sb.WriteString("validation failed at: ")
	for i := len(v.trace)-1; i >= 0; i-- {
		sb.WriteString(v.trace[i])
	}
	return sb.String()
}

func (v *errortrace) attach(msg string) {
	if v != nil {
		v.trace = append(v.trace, msg)
	}
}

func kind(t reflect.Kind) validator {
	return func (value interface{}, exists ...bool) (bool, *errortrace) {
		valid := value != nil && reflect.TypeOf(value).Kind() == t
		if !valid {
			et := errortrace{}
			et.attach(fmt.Sprintf(
				" (expected kind: %s, got value %s of kind %s)",
				t,
				value,
				map[bool]string{
					true: "unknown",
					false: reflect.TypeOf(value).Kind().String(),
				}[value == nil],
			))
			return false, &et
		}
		return true, nil
	}
}

func typ(instance interface{}) validator {
	t := reflect.TypeOf(instance)
	return func (value interface{}, exists ...bool) (bool, *errortrace) {
		valid := value != nil && reflect.TypeOf(value) == t
		if !valid {
			et := errortrace{}
			et.attach(fmt.Sprintf(
				" (expected type: %s, got value %s of type %s)",
				t,
				value,
				map[bool]string{
					true: "unknown",
					false: reflect.TypeOf(value).String(),
				}[value == nil],
			))
			return false, &et
		}
		return true, nil
	}
}

func opt(v validator) validator {
	return func (value interface{}, exists ...bool) (bool, *errortrace) {
		if (len(exists) > 0 && exists[0] == false) {
			return true, nil
		}
		return v(value, exists...)
	}
}

// NB: validation of unspecified fields is skipped
func doc(vm map[string]validator, name ...string) validator {
	return all(
		kind(reflect.Map),
		func (value interface{}, exists ...bool) (bool, *errortrace) {
			for k, v := range vm {
				inner, x := value.(map[string]interface{})[k]
				if valid, et := v(inner, x); !valid {
					if len(name) > 0 {
						et.attach(fmt.Sprintf("%s.%s", name[0], k))
					} else {
						et.attach(fmt.Sprintf(".%s", k))
					}
					return false, et
				}
			}
			return true, nil
		},
	)
}

func list(ev validator) validator {
	return all(
		kind(reflect.Slice),
		func (value interface{}, exists ...bool) (bool, *errortrace) {
			for i, el := range value.([]interface{}) {
				if valid, et := ev(el, true); !valid {
					et.attach(fmt.Sprintf("[%d]", i))
					return false, et
				}
			}
			return true, nil
		},
	)
}

func all(vl ...validator) validator {
	return func (value interface{}, exists ...bool) (bool, *errortrace) {
		for _, v := range vl {
			if valid, et := v(value, exists...); !valid {
				return false, et
			}
		}
		return true, nil
	}
}