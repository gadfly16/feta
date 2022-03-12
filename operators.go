//go:generate pigeon -o parser.go parser.peg

package feta

import "regexp"

type context struct {
	obj  *object
	meta fDict
}

type result struct {
	Obj    *object
	Result fType `json:",omitempty"`
	Err    error `json:",omitempty"`
}

type operator func(*context) []*result

type opProducer interface {
	opProduce(operator) operator
}

type dirOp struct{}

func (opp *dirOp) opProduce(next operator) operator {
	return func(ctx *context) []*result {
		if ctx.obj.dirEntry.IsDir() {
			if next != nil {
				res := []*result{}
				chs, err := ctx.obj.getChildren()
				if err != nil {
					return []*result{{Obj: ctx.obj, Err: fError{err.Error()}}}
				}
				for _, ch := range chs {
					chRes := next(&context{ch, ctx.meta})
					res = append(res, chRes...)
				}
				return res
			}
			return []*result{{Obj: ctx.obj}}
		}
		return nil
	}
}

type patternOp struct {
	rex *regexp.Regexp
}

func (opp *patternOp) opProduce(next operator) operator {
	return func(ctx *context) []*result {
		if opp.rex.Match([]byte(ctx.obj.dirEntry.Name())) {
			if next != nil {
				return next(ctx)
			}
			return []*result{{Obj: ctx.obj}}
		}
		return nil
	}
}

type bootOp struct{}

func (opp *bootOp) opProduce(next operator) operator {
	return func(ctx *context) []*result {
		chs, err := ctx.obj.getChildren()
		if err != nil {
			return []*result{{Obj: ctx.obj, Err: fError{err.Error()}}}
		}
		res := []*result{}
		for _, ch := range chs {
			chRes := next(&context{ch, ctx.meta})
			res = append(res, chRes...)
		}
		return res
	}
}

type rootOp struct{}

func (opp *rootOp) opProduce(next operator) operator {
	return func(_ *context) []*result {
		return next(&context{obj: site})
	}
}

type relativeOp struct {
	count int
}

func (opp *relativeOp) opProduce(next operator) operator {
	return func(ctx *context) []*result {
		anch := ctx.obj
		for i := 1; i < opp.count; i++ {
			anch = anch.parent
		}
		if next != nil {
			return next(&context{anch, ctx.meta})
		}
		return []*result{{Obj: anch}}
	}
}

type recurseOp struct{}

func walk(ctx *context, next operator) []*result {
	chs, err := ctx.obj.getChildren()
	if err != nil {
		return []*result{{Obj: ctx.obj, Err: fError{err.Error()}}}
	}
	res := []*result{}
	for _, ch := range chs {
		if next != nil {
			chRes := next(&context{obj: ch})
			res = append(res, chRes...)
		} else {
			res = append(res, &result{Obj: ch})
		}
		if ch.dirEntry.IsDir() {
			chRes := walk(&context{obj: ch}, next)
			res = append(res, chRes...)
		}
	}
	return res
}

func (opp *recurseOp) opProduce(next operator) operator {
	return func(ctx *context) []*result {
		return walk(ctx, next)
	}
}

type tailOp struct {
	expr exprNode
}

func (opp *tailOp) opProduce(next operator) operator {
	return func(ctx *context) []*result {
		res := []*result{{
			Obj: ctx.obj,
		}}
		ns, err := ctx.obj.getMeta()
		if err != nil {
			res[0].Err = fError{err.Error()}
			return res
		}
		meta, err := opp.expr.eval(ns)
		res[0].Result = meta
		return res
	}
}
