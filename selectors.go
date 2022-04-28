//go:generate pigeon -o parser.go parser.peg

package feta

import (
	"regexp"
)

type context struct {
	obj  *object
	meta fExpr
}

type selector interface {
	sel(*context) fList
	setNext(selector)
}

type dirSel struct {
	count int
	next  selector
}

func (sel *dirSel) setNext(next selector) {
	sel.next = next
}

func (sel *dirSel) sel(ctx *context) fList {
	for i := 1; i < sel.count; i++ {
		if i > 1 {
			ctx.obj = ctx.obj.parent
		}
		proj, err := ctx.obj.getProject()
		if err != nil {
			return fList{fError{err.Error()}}
		}
		if proj == nil {
			return fList{fError{"Couldn't find project for object: " + ctx.obj.fetaPath()}}
		}
		ctx.obj = proj
	}
	if ctx.obj.dirEntry.IsDir() {
		if sel.next != nil {
			switch n := sel.next.(type) {
			case *relSel, *recurseSel:
				return n.sel(ctx)
			}
			res := fList{}
			chs, err := ctx.obj.getChildren()
			if err != nil {
				return fList{fError{err.Error() + " at " + ctx.obj.fetaPath()}}
			}
			for _, ch := range chs {
				chRes := sel.next.sel(&context{ch, ctx.meta})
				res = append(res, chRes...)
			}
			return res
		}
		return fList{ctx.obj}
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

func (sel *patternSel) sel(ctx *context) fList {
	if sel.rex.Match([]byte(ctx.obj.dirEntry.Name())) {
		if sel.next != nil {
			return sel.next.sel(ctx)
		}
		return fList{ctx.obj}
	}
	return nil
}

type rootSel struct {
	next selector
}

func (sel *rootSel) setNext(next selector) {
	sel.next = next
}

func (sel *rootSel) sel(_ *context) fList {
	return sel.next.sel(&context{obj: site})
}

type relSel struct {
	count int
	next  selector
}

func (sel *relSel) setNext(next selector) {
	sel.next = next
}

func (sel *relSel) sel(ctx *context) fList {
	anch := ctx.obj
	for i := 1; i < sel.count; i++ {
		anch = anch.parent
		if anch == nil {
			return fList{fError{"Invalid relative reference from " + ctx.obj.fetaPath()}}
		}
	}
	if sel.next != nil {
		return sel.next.sel(&context{anch, ctx.meta})
	}
	return fList{anch}
}

type recurseSel struct {
	next selector
}

func (sel *recurseSel) setNext(next selector) {
	sel.next = next
}

func walk(ctx *context, next selector) fList {
	chs, err := ctx.obj.getChildren()
	if err != nil {
		return fList{fError{err.Error() + " at " + ctx.obj.fetaPath()}}
	}
	res := fList{}
	for _, ch := range chs {
		if next != nil {
			chRes := next.sel(&context{obj: ch})
			res = append(res, chRes...)
		} else {
			res = append(res, ch)
		}
		if ch.dirEntry.IsDir() {
			chRes := walk(&context{obj: ch}, next)
			res = append(res, chRes...)
		}
	}
	return res
}

func (sel *recurseSel) sel(ctx *context) fList {
	return walk(ctx, sel.next)
}

type tailSel struct {
	expr fExpr
}

func (sel *tailSel) setNext(next selector) {
}

func (sel *tailSel) sel(ctx *context) fList {
	ns, err := ctx.obj.getMeta()
	if err != nil {
		return fList{fError{err.Error() + " at " + ctx.obj.fetaPath()}}
	}
	res := fDict{"Obj": ctx.obj}
	if sel.expr == nil {
		for k, v := range procedurals {
			pr := v.eval(ctx)
			if fErr, ok := pr.(fError); ok {
				return fList{fErr}
			}
			ns[k] = pr
		}
		res["Value"] = ns
		return fList{res}
	}
	meta := sel.expr.eval(&context{obj: ctx.obj, meta: ns})
	if fErr, ok := meta.(fError); ok {
		return fList{fErr}
	}
	res["Value"] = meta
	return fList{res}
}

type filterSel struct {
	expr fExpr
	next selector
}

func (sel *filterSel) setNext(next selector) {
	sel.next = next
}

func (sel *filterSel) get(ctx *context) fList {
	ns, err := ctx.obj.getMeta()
	if err != nil {
		return fList{fError{err.Error() + " at " + ctx.obj.fetaPath()}}
	}
	value := sel.expr.eval(&context{obj: ctx.obj, meta: ns})
	if fErr, ok := value.(fError); ok {
		return fList{fErr}
	}
	if value != nil && boolVal(value) {
		if sel.next != nil {
			return sel.next.sel(ctx)
		}
		return fList{ctx.obj}
	}
	return fList{}
}
