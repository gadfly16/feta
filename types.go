package feta

import (
	"errors"
	"time"
)

type Meta interface {
}

type MNumber float64
type MText string
type MTime time.Time
type MExp string
type MMap map[string]Meta
type MList []Meta

func (t MTime) String() string {
	return time.Time(t).Format(time.RFC3339)
}

func (t MTime) MarshalJSON() ([]byte, error) {
	return []byte("\"T" + t.String() + "\""), nil
}

func (x MExp) MarshalJSON() ([]byte, error) {
	return []byte("\"=" + x + "\""), nil
}

func TypeConvert(i interface{}) (Meta, error) {
	switch t := i.(type) {
	case map[string]interface{}:
		m := make(MMap)
		var err error
		for k, v := range t {
			m[k], err = TypeConvert(v)
			if err != nil {
				return MNumber(0), err
			}
		}
		return m, nil
	case float64:
		return MNumber(t), nil
	case string:
		return MText(t), nil
	default:
		return MNumber(0), errors.New("Unknown type found.")
	}
}
