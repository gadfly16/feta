package feta

type resolver interface {
	setNextAndRaw(resolver, bool)
	resolve(*context, fExpr) fExpr
}

type valueRes struct {
	expr fExpr
	next resolver
	raw  bool
}

func (node *valueRes) setNextAndRaw(next resolver, raw bool) {
	node.next = next
	node.raw = raw
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

type contextRes struct {
	next resolver
	raw  bool
}

func (node *contextRes) setNextAndRaw(next resolver, raw bool) {
	node.next = next
	node.raw = raw
}

func (node *contextRes) eval(ctx *context) fExpr {
	return node.resolve(ctx, ctx.meta)
}

func (node *contextRes) resolve(ctx *context, ns fExpr) fExpr {
	if node.next == nil {
		if node.raw {
			return ctx.meta
		}
		return ctx.meta.eval(ctx)
	}
	return node.next.resolve(ctx, ctx.meta)
}

type indexRes struct {
	expr fExpr
	next resolver
	raw  bool
}

func (node *indexRes) setNextAndRaw(next resolver, raw bool) {
	node.next = next
	node.raw = raw
}

func (node *indexRes) resolve(ctx *context, ns fExpr) fExpr {
	index := node.expr.eval(ctx)
	if fErr, ok := index.(fError); ok {
		return fErr
	}
	var res fExpr
	var exists bool
	switch v := ns.(type) {
	case fDict:
		i, isStr := index.(fString)
		if !isStr {
			return fError{"Dicts can only be indexed with strings."}
		}
		res, exists = v[string(i)]
		if !exists {
			return fNone{}
		}
	case fList:
		i, isNum := index.(fNumber)
		if !isNum {
			return fError{"Lists can only be indexed with numbers."}
		}
		ii := int(i)
		if ii > len(v)-1 || ii < 0 {
			return fError{"Index out of range."}
		}
		res = v[ii]
	default:
		return fError{"Only lists and dicts can be indexed."}
	}
	if node.next == nil {
		if node.raw {
			return res
		}
		return res.eval(ctx)
	}
	if node.raw {
		switch r := res.(type) {
		case fDict, fList:
			return node.next.resolve(ctx, r)
		}
	}
	ns = res.eval(ctx)
	if fErr, ok := ns.(fError); ok {
		return fErr
	}
	return node.next.resolve(ctx, ns)
}

type attribRes struct {
	identifier string
	next       resolver
	raw        bool
}

func (node *attribRes) setNextAndRaw(next resolver, raw bool) {
	node.next = next
	node.raw = raw
}

func (node *attribRes) eval(ctx *context) fExpr {
	return node.resolve(ctx, ctx.meta)
}

func (node *attribRes) resolve(ctx *context, ns fExpr) fExpr {
	if node.identifier == "" {
		if node.next == nil {
			if node.raw {
				return ns
			}
			return ns.eval(ctx)
		}
		return node.next.resolve(ctx, ns)
	}
	switch t := ns.(type) {
	case fDict:
		res, exists := t[node.identifier]
		if exists {
			if node.next == nil {
				if node.raw {
					return res
				}
				return res.eval(ctx)
			}
			if node.raw {
				switch r := res.(type) {
				case fDict, fList:
					return node.next.resolve(ctx, r)
				}
			}
			ns := res.eval(ctx)
			if fErr, ok := ns.(fError); ok {
				return fErr
			}
			return node.next.resolve(ctx, ns)
		}
		return fNone{}
	}
	return fError{"Trying to access name in non-object type."}
}

type compareNode struct {
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

func (node *compareNode) eval(ctx *context) fExpr {
	left := node.left.eval(ctx)
	if fErr, ok := left.(fError); ok {
		return fErr
	}
	right := node.right.eval(ctx)
	if fErr, ok := right.(fError); ok {
		return fErr
	}
	switch l := left.(type) {
	case fNone:
		_, same := right.(fNone)
		switch node.op {
		case EQ:
			return fBool(same)
		case NEQ:
			return fBool(!same)
		}
		return fError{"None is not orderable."}
	case fBool:
		r, same := right.(fBool)
		if !same {
			switch node.op {
			case EQ:
				return fBool(false)
			case NEQ:
				return fBool(true)
			}
			return fError{"Only '==' and '!=' is supported between different types."}
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
			switch node.op {
			case EQ:
				return fBool(false)
			case NEQ:
				return fBool(true)
			}
			return fError{"Only '==' and '!=' is supported between different types."}
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
			switch node.op {
			case EQ:
				return fBool(false)
			case NEQ:
				return fBool(true)
			}
			return fError{"Only '==' and '!=' is supported between different types."}
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

type compoundNode struct {
	expr fExpr
}

func (node *compoundNode) eval(ctx *context) fExpr {
	return node.expr.eval(ctx)
}

type notNode struct {
	expr fExpr
}

func (node *notNode) eval(ctx *context) fExpr {
	return !boolVal(node.expr.eval(ctx))
}

type queryNode struct {
	sel   selector
	multi bool
	tail  bool
}

func (node *queryNode) eval(ctx *context) fExpr {
	res := node.sel.sel(ctx)
	if !node.multi && len(res) != 0 {
		if node.tail {
			if err, ok := res[0].(fError); ok {
				return err
			}
			return res[0].(fDict)["Value"]
		}
		return res[0]
	}
	return res
}
