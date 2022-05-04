package feta

import (
	"strconv"
	"strings"
)

const (
	indentWidth = 2
	startSize   = 4096
)

type fNode interface {
	marshal(*mshState)
}

type mshState struct {
	res    []byte
	pretty bool
	indent int
}

func marshal(node fNode, pretty bool) []byte {
	res := make([]byte, 0, startSize)
	st := &mshState{res, pretty, 0}
	node.marshal(st)
	st.res = append(st.res, '\n')
	return st.res
}

func (value *object) marshal(st *mshState) {
	st.res = append(st.res, "`"+value.fetaPath()+"`"...)
}

func (value fError) marshal(st *mshState) {
	st.res = append(st.res, "error{\""+value.msg+"\"}"...)
}

func (value fBool) marshal(st *mshState) {
	if value {
		st.res = append(st.res, "true"...)
		return
	}
	st.res = append(st.res, "false"...)
}

func (value fNone) marshal(st *mshState) {
	st.res = append(st.res, "none"...)
}

func (value fNumber) marshal(st *mshState) {
	st.res = append(st.res, strconv.FormatFloat(float64(value), 'f', -1, 64)...)
}

func (value fString) marshal(st *mshState) {
	st.res = append(st.res, "\""+string(value)+"\""...)
}

func (value fDict) marshal(st *mshState) {
	var ind string
	if st.pretty {
		st.res = append(st.res, "{\n"...)
		st.indent++
	} else {
		st.res = append(st.res, '{')
	}
	ind = strings.Repeat(" ", st.indent*indentWidth)
	last := len(value) - 1
	i := 0
	for k, v := range value {
		if st.pretty {
			st.res = append(st.res, ind...)
		}
		st.res = append(st.res, (k + ": ")...)
		v.(fNode).marshal(st)
		if i < last {
			if st.pretty {
				st.res = append(st.res, ",\n"...)
			} else {
				st.res = append(st.res, ',')
			}
		}
		i++
	}
	if st.pretty {
		st.indent--
		ind = strings.Repeat(" ", st.indent*indentWidth)
		st.res = append(st.res, ("\n" + ind + "}")...)
	} else {
		st.res = append(st.res, '}')
	}
}

func (value fList) marshal(st *mshState) {
	var ind string
	if st.pretty {
		st.res = append(st.res, "[\n"...)
		st.indent++
	} else {
		st.res = append(st.res, '[')
	}
	ind = strings.Repeat(" ", st.indent*indentWidth)
	last := len(value) - 1
	for i, v := range value {
		if st.pretty {
			st.res = append(st.res, ind...)
		}
		v.(fNode).marshal(st)
		if i < last {
			if st.pretty {
				st.res = append(st.res, ",\n"...)
			} else {
				st.res = append(st.res, ',')
			}
		}
	}
	if st.pretty {
		st.indent--
		ind = strings.Repeat(" ", st.indent*indentWidth)
		st.res = append(st.res, ("\n" + ind + "]")...)
	} else {
		st.res = append(st.res, ']')
	}
}

func (node *multNode) marshal(st *mshState) {
	node.left.(fNode).marshal(st)
	st.res = append(st.res, node.op)
	node.right.(fNode).marshal(st)
}

func (node *addNode) marshal(st *mshState) {
	node.left.(fNode).marshal(st)
	st.res = append(st.res, node.op)
	node.right.(fNode).marshal(st)
}

func (node *compoundNode) marshal(st *mshState) {
	st.res = append(st.res, '(')
	node.expr.(fNode).marshal(st)
	st.res = append(st.res, ')')
}

func (node *notNode) marshal(st *mshState) {
	st.res = append(st.res, '!')
	node.expr.(fNode).marshal(st)
}

func (node *objProc) marshal(st *mshState) {
	st.res = append(st.res, "objProc{}"...)
}
