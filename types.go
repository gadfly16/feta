package feta

import (
	"encoding/json"
	"errors"
	"time"
)

type fType interface {
}

type fBool bool
type fNumber float64
type fString string
type fTime time.Time
type fExpr string
type fDict map[string]fType
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

func (t fTime) String() string {
	return time.Time(t).Format(time.RFC3339)
}

func (t fTime) MarshalJSON() ([]byte, error) {
	return []byte("\"T" + t.String() + "\""), nil
}

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
