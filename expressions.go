package feta

type resolver interface {
	setNext(resolver)
	resolve(*context, fExpr) fExpr
}

type valueRes struct {
	expr fExpr
	next resolver
}

func (node *valueRes) setNext(next resolver) {
	node.next = next
}

func (node *valueRes) eval(ctx *context) fExpr {
	value := node.expr.eval(ctx)
	if fErr, ok := value.(fError); ok {
		return fErr
	}
	return node.next.resolve(ctx, value)
}

func (node *valueRes) resolve(ctx *context, ns fExpr) fExpr {
	return nil
}

type indexRes struct {
	expr fExpr
	next resolver
}

func (node *indexRes) setNext(next resolver) {
	node.next = next
}

func (node *indexRes) resolve(ctx *context, ns fExpr) fExpr {
	index := node.expr.eval(ctx)
	if fErr, ok := index.(fError); ok {
		return fErr
	}
	switch v := ns.(type) {
	case fDict:
		i, isStr := index.(fString)
		if !isStr {
			return fError{"Dicts can only be indexed with strings."}
		}
		res, exists := v[string(i)]
		if exists {
			if node.next == nil {
				return res.eval(ctx)
			}
			ns := res.eval(ctx)
			if fErr, ok := ns.(fError); ok {
				return fErr
			}
			return node.next.resolve(ctx, ns)
		}
		return nil
	case fList:
		i, isNum := index.(fNumber)
		if !isNum {
			return fError{"Lists can only be indexed with numbers."}
		}
		ii := int(i)
		if ii > len(v)-1 || ii < 0 {
			return fError{"Index out of range."}
		}
		res := v[ii]
		if node.next == nil {
			return res.eval(ctx)
		}
		ns := res.eval(ctx)
		if fErr, ok := ns.(fError); ok {
			return fErr
		}
		return node.next.resolve(ctx, ns)
	}
	return fError{"Only lists and dicts can be indexed."}
}

type attribRes struct {
	identifier string
	next       resolver
}

func (node *attribRes) setNext(next resolver) {
	node.next = next
}

func (node *attribRes) eval(ctx *context) fExpr {
	return node.resolve(ctx, ctx.meta)
}

func (node *attribRes) resolve(ctx *context, ns fExpr) fExpr {
	switch t := ns.(type) {
	case fDict:
		res, exists := t[node.identifier]
		if exists {
			if node.next == nil {
				return res.eval(ctx)
			}
			ns := res.eval(ctx)
			if fErr, ok := ns.(fError); ok {
				return fErr
			}
			return node.next.resolve(ctx, ns)
		}
		return nil
	}
	return fError{"Trying to access name in non-object type."}
}

type compNode struct {
	op    byte
	left  fExpr
	right fExpr
}

const (
	EQ = iota
	NEQ
	LEEQ
	GREQ
	LE
	GR
)

func (node *compNode) eval(ctx *context) fExpr {
	left := node.left.eval(ctx)
	if fErr, ok := left.(fError); ok {
		return fErr
	}
	right := node.right.eval(ctx)
	if fErr, ok := right.(fError); ok {
		return fErr
	}
	switch l := left.(type) {
	case nil:
		if right == nil {
			return fBool(true)
		}
		return fBool(false)
	case fBool:
		r, same := right.(fBool)
		if !same {
			return fBool(false)
		}
		switch node.op {
		case EQ:
			return fBool(l == r)
		case NEQ:
			return fBool(l != r)
		}
		return fError{"Booleans are not orderable."}
	case fNumber:
		r, same := right.(fNumber)
		if !same {
			return fBool(false)
		}
		switch node.op {
		case EQ:
			return fBool(l == r)
		case NEQ:
			return fBool(l != r)
		case LEEQ:
			return fBool(l <= r)
		case GREQ:
			return fBool(l >= r)
		case LE:
			return fBool(l < r)
		case GR:
			return fBool(l > r)
		}
	case fString:
		r, same := right.(fString)
		if !same {
			return fBool(false)
		}
		switch node.op {
		case EQ:
			return fBool(l == r)
		case NEQ:
			return fBool(l != r)
		case LEEQ:
			return fBool(l <= r)
		case GREQ:
			return fBool(l >= r)
		case LE:
			return fBool(l < r)
		case GR:
			return fBool(l > r)
		}
	}
	return fError{"Only numbers and strings can be compared."}
}

type addNode struct {
	op    byte
	left  fExpr
	right fExpr
}

func (node *addNode) eval(ctx *context) fExpr {
	left := node.left.eval(ctx)
	if fErr, ok := left.(fError); ok {
		return fErr
	}
	right := node.right.eval(ctx)
	if fErr, ok := right.(fError); ok {
		return fErr
	}
	switch l := left.(type) {
	case fNumber:
		r, same := right.(fNumber)
		if !same {
			return fError{"Nubers can only be added to numbers."}
		}
		if node.op == '+' {
			return l + r
		}
		return l - r
	case fString:
		r, same := right.(fString)
		if !same {
			return fError{"Strings can only be added to strings."}
		}
		if node.op == '+' {
			return l + r
		}
		return fError{"Strings can not be subtracted from strings."}
	}
	return fError{"Only numbers and strings can be added."}
}

type multNode struct {
	op    byte
	left  fExpr
	right fExpr
}

func (node *multNode) eval(ctx *context) fExpr {
	left := node.left.eval(ctx)
	if fErr, ok := left.(fError); ok {
		return fErr
	}
	right := node.right.eval(ctx)
	if fErr, ok := right.(fError); ok {
		return fErr
	}
	switch l := left.(type) {
	case fNumber:
		r, same := right.(fNumber)
		if !same {
			return fError{"Nubers can only be multiplied by numbers."}
		}
		if node.op == '*' {
			return l * r
		}
		return l / r
	}
	return fError{"Only numbers can be multiplied."}
}

type andNode struct {
	left  fExpr
	right fExpr
}

func (node *andNode) eval(ctx *context) fExpr {
	left := node.left.eval(ctx)
	if fErr, ok := left.(fError); ok {
		return fErr
	}
	right := node.right.eval(ctx)
	if fErr, ok := right.(fError); ok {
		return fErr
	}
	return boolVal(left) && boolVal(right)
}

type orNode struct {
	left  fExpr
	right fExpr
}

func (node *orNode) eval(ctx *context) fExpr {
	left := node.left.eval(ctx)
	if fErr, ok := left.(fError); ok {
		return fErr
	}
	right := node.right.eval(ctx)
	if fErr, ok := right.(fError); ok {
		return fErr
	}
	return boolVal(left) || boolVal(right)
}
