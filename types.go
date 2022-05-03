package feta

type fExpr interface {
	eval(*context) fExpr
}

type (
	fBool   bool
	fNumber float64
	fString string
	fDict   map[string]fExpr
	fList   []fExpr
	fError  struct{ msg string }
	fNone   struct{}
)

func boolVal(node fExpr) fBool {
	switch n := node.(type) {
	case fBool:
		return n
	case fNone:
		return fBool(false)
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

func (value fBool) eval(ctx *context) fExpr {
	return value
}

func (value fNone) eval(ctx *context) fExpr {
	return value
}

func (value fNumber) eval(ctx *context) fExpr {
	return value
}

func (value fString) eval(ctx *context) fExpr {
	return value
}

func (value fDict) eval(ctx *context) fExpr {
	for k, elm := range value {
		res := elm.eval(ctx)
		if fErr, ok := res.(fError); ok {
			return fErr
		}
		value[k] = res
	}
	return value
}

func (value fList) eval(ctx *context) fExpr {
	for i, elm := range value {
		res := elm.eval(ctx)
		if fErr, ok := res.(fError); ok {
			return fErr
		}
		value[i] = res
	}
	return value
}

func (value fError) eval(ctx *context) fExpr {
	return value
}

func (e fError) Error() string {
	return e.msg
}
