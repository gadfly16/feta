package feta

import (
	"encoding/json"
	"errors"
)

type fType interface {
	boolVal() fBool
}

type fBool bool

func (value fBool) boolVal() fBool {
	return value
}

func (value fBool) eval(ctx *context) fType {
	return value
}

type fNumber float64

func (value fNumber) boolVal() fBool {
	return value != 0
}

func (value fNumber) eval(ctx *context) fType {
	return value
}

type fString string

func (value fString) boolVal() fBool {
	return value != ""
}

func (value fString) eval(ctx *context) fType {
	return value
}

type fDict map[string]expression

func (value fDict) boolVal() fBool {
	return len(value) != 0
}

func (value fDict) eval(ctx *context) fType {
	for k, elm := range value {
		res := elm.eval(ctx)
		if fErr, ok := res.(fError); ok {
			return fErr
		}
		value[k] = res.(expression)
	}
	return value
}

type fList []expression

func (value fList) boolVal() fBool {
	return len(value) != 0
}

func (value fList) eval(ctx *context) fType {
	for i, elm := range value {
		res := elm.eval(ctx)
		if fErr, ok := res.(fError); ok {
			return fErr
		}
		value[i] = res.(expression)
	}
	return value
}

type fError struct {
	msg string
}

func (value fError) boolVal() fBool {
	return len(value.msg) != 0
}

func (value fError) eval(ctx *context) fType {
	return value
}

func (e fError) Error() string {
	return e.msg
}

func (e fError) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.msg)
}

func typeConvert(i interface{}) (fType, error) {
	switch t := i.(type) {
	case map[string]interface{}:
		m := make(fDict)
		for k, v := range t {
			fv, err := typeConvert(v)
			if err != nil {
				return nil, err
			}
			m[k] = fv.(expression)
		}
		return m, nil
	case []interface{}:
		l := make(fList, 0)
		for _, v := range t {
			fv, err := typeConvert(v)
			if err != nil {
				return nil, err
			}
			l = append(l, fv.(expression))
		}
		return l, nil
	case float64:
		return fNumber(t), nil
	case string:
		return fString(t), nil
	}
	return nil, errors.New("Unknown type found.")
}
