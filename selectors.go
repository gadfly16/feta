//go:generate pigeon -o parser.go parser.peg

package feta

import (
	"regexp"
)

type context struct {
	obj  *object
	meta fType
}

type result struct {
	Obj    *object `json:",omitempty"`
	Result fType   `json:",omitempty"`
	Error  error   `json:",omitempty"`
}

type selector interface {
	get(*context) []*result
	setNext(selector)
}

type dirSel struct {
	next selector
}

func (sel *dirSel) setNext(next selector) {
	sel.next = next
}

func (sel *dirSel) get(ctx *context) []*result {
	if ctx.obj.dirEntry.IsDir() {
		if sel.next != nil {
			switch n := sel.next.(type) {
			case *relSel, *recurseSel:
				return n.get(ctx)
			}
			res := []*result{}
			chs, err := ctx.obj.getChildren()
			if err != nil {
				return []*result{{Obj: ctx.obj, Error: fError{err.Error()}}}
			}
			for _, ch := range chs {
				chRes := sel.next.get(&context{ch, ctx.meta})
				res = append(res, chRes...)
			}
			return res
		}
		return []*result{{Obj: ctx.obj}}
	}
	return nil
}

type patternSel struct {
	rex  *regexp.Regexp
	next selector
}

func (sel *patternSel) setNext(next selector) {
	sel.next = next
}

func (sel *patternSel) get(ctx *context) []*result {
	if sel.rex.Match([]byte(ctx.obj.dirEntry.Name())) {
		if sel.next != nil {
			return sel.next.get(ctx)
		}
		return []*result{{Obj: ctx.obj}}
	}
	return nil
}

type rootSel struct {
	next selector
}

func (sel *rootSel) setNext(next selector) {
	sel.next = next
}

func (sel *rootSel) get(_ *context) []*result {
	return sel.next.get(&context{obj: site})
}

type relSel struct {
	count int
	next  selector
}

func (sel *relSel) setNext(next selector) {
	sel.next = next
}

func (sel *relSel) get(ctx *context) []*result {
	anch := ctx.obj
	for i := 1; i < sel.count; i++ {
		anch = anch.parent
		if anch == nil {
			return []*result{{Error: fError{"Invalid relative reference from " + ctx.obj.fetaPath()}}}
		}
	}
	if sel.next != nil {
		return sel.next.get(&context{anch, ctx.meta})
	}
	return []*result{{Obj: anch}}
}

type recurseSel struct {
	next selector
}

func (sel *recurseSel) setNext(next selector) {
	sel.next = next
}

func walk(ctx *context, next selector) []*result {
	chs, err := ctx.obj.getChildren()
	if err != nil {
		return []*result{{Obj: ctx.obj, Error: fError{err.Error()}}}
	}
	res := []*result{}
	for _, ch := range chs {
		if next != nil {
			chRes := next.get(&context{obj: ch})
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

func (sel *recurseSel) get(ctx *context) []*result {
	return walk(ctx, sel.next)
}

type tailSel struct {
	expr expression
}

func (sel *tailSel) setNext(next selector) {
}

func (sel *tailSel) get(ctx *context) []*result {
	res := []*result{{
		Obj: ctx.obj,
	}}
	ns, err := ctx.obj.getMeta()
	if err != nil {
		res[0].Error = fError{err.Error()}
		return res
	}
	if sel.expr == nil {
		for k, v := range procedurals {
			pr, err := v.eval(ctx)
			if err != nil {
				res[0].Error = fError{err.Error()}
				return res
			}
			ns.(fDict)[k] = pr.(expression)
		}
		res[0].Result = ns
		return res
	}
	meta, err := sel.expr.eval(&context{obj: ctx.obj, meta: ns})
	if err != nil {
		res[0].Error = fError{err.Error()}
		return res
	}
	res[0].Result = meta
	return res
}

type filterSel struct {
	expr expression
	next selector
}

func (sel *filterSel) setNext(next selector) {
	sel.next = next
}

func (sel *filterSel) get(ctx *context) []*result {
	res := []*result{{
		Obj: ctx.obj,
	}}
	ns, err := ctx.obj.getMeta()
	if err != nil {
		res[0].Error = fError{err.Error()}
		return res
	}
	value, err := sel.expr.eval(&context{obj: ctx.obj, meta: ns})
	if err != nil {
		res[0].Error = fError{err.Error()}
		return res
	}
	if value != nil && value.boolVal() {
		if sel.next != nil {
			return sel.next.get(ctx)
		}
		return res
	}
	return []*result{}
}
