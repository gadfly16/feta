package feta

import (
	"encoding/json"
	"errors"
)

type fNode interface {
	eval(*context) fNode
}

type (
	fBool   bool
	fNumber float64
	fString string
	fDict   map[string]fNode
	fList   []fNode
)

type fError struct {
	msg string
}

func boolVal(node fNode) fBool {
	switch n := node.(type) {
	case fBool:
		return n
	case fNumber:
		return n != 0
	case fString:
		return len(n) != 0
	case fDict:
		return len(n) != 0
	case fList:
		return len(n) != 0
	case fError:
		return len(n.msg) != 0
	}
	return fBool(true)
}

func (value fBool) eval(ctx *context) fNode {
	return value
}

func (value fNumber) eval(ctx *context) fNode {
	return value
}

func (value fString) eval(ctx *context) fNode {
	return value
}

func (value fDict) eval(ctx *context) fNode {
	for k, elm := range value {
		res := elm.eval(ctx)
		if fErr, ok := res.(fError); ok {
			return fErr
		}
		value[k] = res
	}
	return value
}

func (value fList) eval(ctx *context) fNode {
	for i, elm := range value {
		res := elm.eval(ctx)
		if fErr, ok := res.(fError); ok {
			return fErr
		}
		value[i] = res
	}
	return value
}

func (value fError) eval(ctx *context) fNode {
	return value
}

func (e fError) Error() string {
	return e.msg
}

func (e fError) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.msg)
}

func typeConvert(i interface{}) (fNode, error) {
	switch t := i.(type) {
	case map[string]interface{}:
		m := make(fDict)
		for k, v := range t {
			fv, err := typeConvert(v)
			if err != nil {
				return nil, err
			}
			m[k] = fv
		}
		return m, nil
	case []interface{}:
		l := make(fList, 0)
		for _, v := range t {
			fv, err := typeConvert(v)
			if err != nil {
				return nil, err
			}
			l = append(l, fv)
		}
		return l, nil
	case float64:
		return fNumber(t), nil
	case string:
		return fString(t), nil
	}
	return nil, errors.New("Unknown type found.")
}
