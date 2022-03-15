package feta

import "errors"

type exprNode interface {
	eval(*context) (fType, error)
}

type keyNode struct {
	key  string
	next exprNode
}

func (node *keyNode) eval(ctx *context) (fType, error) {
	switch t := ctx.meta.(type) {
	case fDict:
		res, exists := t[node.key]
		if exists {
			if node.next == nil {
				return res, nil
			}
			return node.next.eval(&context{obj: ctx.obj, meta: res})
		}
		return nil, nil
	}
	return nil, errors.New("Trying to access name in non-object type.")
}

type numberNode struct {
	value fNumber
}

func (node *numberNode) eval(ctx *context) (fType, error) {
	return node.value, nil
}

const (
	EQ = iota
	LEEQ
	GREQ
	LE
	GR
)

type compNode struct {
	op    byte
	left  exprNode
	right exprNode
}

func (node *compNode) eval(ctx *context) (fType, error) {
	left, err := node.left.eval(ctx)
	if err != nil {
		return nil, err
	}
	right, err := node.right.eval(ctx)
	if err != nil {
		return nil, err
	}
	switch l := left.(type) {
	case fNumber:
		r, same := right.(fNumber)
		if !same {
			return nil, errors.New("Nubers can only be compared to numbers.")
		}
		switch node.op {
		case EQ:
			return fBool(l == r), nil
		case LEEQ:
			return fBool(l <= r), nil
		case GREQ:
			return fBool(l >= r), nil
		case LE:
			return fBool(l < r), nil
		case GR:
			return fBool(l > r), nil
		}
	}
	return nil, errors.New("Only numbers can be compared.")
}
