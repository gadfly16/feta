package feta

import "errors"

type expression interface {
	eval(*context) (fType, error)
}

type resolver interface {
	setNext(expression)
}

type valueRes struct {
	expr expression
	next expression
}

func (node *valueRes) setNext(next expression) {
	node.next = next
}

func (node *valueRes) eval(ctx *context) (fType, error) {
	value, err := node.expr.eval(ctx)
	if err != nil {
		return nil, err
	}
	// if node.next != nil {
	return node.next.eval(&context{obj: ctx.obj, meta: value})
	// // }
	// return value, nil
}

type indexRes struct {
	expr expression
	next expression
}

func (node *indexRes) setNext(next expression) {
	node.next = next
}

func (node *indexRes) eval(ctx *context) (fType, error) {
	index, err := node.expr.eval(ctx)
	if err != nil {
		return nil, err
	}
	switch v := ctx.meta.(type) {
	case fDict:
		i, isStr := index.(fString)
		if !isStr {
			return nil, errors.New("Dicts can only be indexed with strings.")
		}
		res, exists := v[string(i)]
		if exists {
			if node.next == nil {
				return res, nil
			}
			return node.next.eval(&context{obj: ctx.obj, meta: res})
		}
		return nil, nil
	case fList:
		i, isNum := index.(fNumber)
		if !isNum {
			return nil, errors.New("Lists can only be indexed with numbers.")
		}
		ii := int(i)
		if ii > len(v)-1 || ii < 0 {
			return nil, errors.New("Index out of range.")
		}
		res := v[ii]
		if node.next == nil {
			return res, nil
		}
		return node.next.eval(&context{obj: ctx.obj, meta: res})
	}
	return nil, errors.New("Only lists and dicts can be indexed.")
}

type attribRes struct {
	identifier string
	next       expression
}

func (node *attribRes) setNext(next expression) {
	node.next = next
}

func (node *attribRes) eval(ctx *context) (fType, error) {
	switch t := ctx.meta.(type) {
	case fDict:
		res, exists := t[node.identifier]
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

type compNode struct {
	op    byte
	left  expression
	right expression
}

const (
	EQ = iota
	NEQ
	LEEQ
	GREQ
	LE
	GR
)

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
	case fBool:
		r, same := right.(fBool)
		if !same {
			return nil, errors.New("Booleans can only be compared to booleans. Yet..")
		}
		switch node.op {
		case EQ:
			return fBool(l == r), nil
		case NEQ:
			return fBool(l != r), nil
		}
		return nil, errors.New("Booleans are not orderable.")
	case fNumber:
		r, same := right.(fNumber)
		if !same {
			return nil, errors.New("Numbers can only be compared to numbers.")
		}
		switch node.op {
		case EQ:
			return fBool(l == r), nil
		case NEQ:
			return fBool(l != r), nil
		case LEEQ:
			return fBool(l <= r), nil
		case GREQ:
			return fBool(l >= r), nil
		case LE:
			return fBool(l < r), nil
		case GR:
			return fBool(l > r), nil
		}
	case fString:
		r, same := right.(fString)
		if !same {
			return nil, errors.New("Strings can only be compared to strings.")
		}
		switch node.op {
		case EQ:
			return fBool(l == r), nil
		case NEQ:
			return fBool(l != r), nil
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
	return nil, errors.New("Only numbers and strings can be compared.")
}

type addNode struct {
	op    byte
	left  expression
	right expression
}

func (node *addNode) eval(ctx *context) (fType, error) {
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
			return nil, errors.New("Nubers can only be added to numbers.")
		}
		if node.op == '+' {
			return l + r, nil
		}
		return l - r, nil
	case fString:
		r, same := right.(fString)
		if !same {
			return nil, errors.New("Strings can only be added to strings.")
		}
		if node.op == '+' {
			return l + r, nil
		}
		return nil, errors.New("Strings can not be subtracted from strings.")
	}
	return nil, errors.New("Only numbers and strings can be added.")
}

type multNode struct {
	op    byte
	left  expression
	right expression
}

func (node *multNode) eval(ctx *context) (fType, error) {
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
			return nil, errors.New("Nubers can only be multiplied by numbers.")
		}
		if node.op == '*' {
			return l * r, nil
		}
		return l / r, nil
	}
	return nil, errors.New("Only numbers can be multiplied.")
}

type andNode struct {
	left  expression
	right expression
}

func (node *andNode) eval(ctx *context) (fType, error) {
	left, err := node.left.eval(ctx)
	if err != nil {
		return nil, err
	}
	right, err := node.right.eval(ctx)
	if err != nil {
		return nil, err
	}
	return left.boolVal() && right.boolVal(), nil
}

type orNode struct {
	left  expression
	right expression
}

func (node *orNode) eval(ctx *context) (fType, error) {
	left, err := node.left.eval(ctx)
	if err != nil {
		return nil, err
	}
	right, err := node.right.eval(ctx)
	if err != nil {
		return nil, err
	}
	return left.boolVal() || right.boolVal(), nil
}
