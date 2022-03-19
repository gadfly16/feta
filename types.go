package feta

import (
	"encoding/json"
	"errors"
	"time"
)

type fType interface {
	boolVal() fBool
}

type fBool bool

func (value fBool) boolVal() fBool {
	return value
}

func (value fBool) eval(ctx *context) (fType, error) {
	return value, nil
}

type fNumber float64

func (value fNumber) boolVal() fBool {
	return value != 0
}

func (value fNumber) eval(ctx *context) (fType, error) {
	return value, nil
}

type fString string

func (value fString) boolVal() fBool {
	return value != ""
}

func (value fString) eval(ctx *context) (fType, error) {
	return value, nil
}

type fDict map[string]fType

func (value fDict) boolVal() fBool {
	return len(value) != 0
}

type fList []fType

type fError struct {
	msg string
}

func (e fError) Error() string {
	return e.msg
}

func (e fError) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.msg)
}

type fTime time.Time

func (t fTime) String() string {
	return time.Time(t).Format(time.RFC3339)
}

func (t fTime) MarshalJSON() ([]byte, error) {
	return []byte("\"T" + t.String() + "\""), nil
}

type fExpr string

func (x fExpr) MarshalJSON() ([]byte, error) {
	return []byte("\"=" + x + "\""), nil
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
			m[k] = fv
		}
		return m, nil
	case []interface{}:
		l := make(fList, 1)
		for _, v := range t {
			fv, err := typeConvert(v)
			if err != nil {
				return nil, err
			}
			l = append(l, fv)
		}
	case float64:
		return fNumber(t), nil
	case string:
		return fString(t), nil
	}
	return nil, errors.New("Unknown type found.")
}
