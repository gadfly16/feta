//go:generate pigeon -o parser.go parser.peg

package feta

import "regexp"

type operator func(*object) ([]*object, error)

type opProducer interface {
	opProduce(operator) operator
}

type dirOp struct{}

func (opp *dirOp) opProduce(next operator) operator {
	return func(start *object) ([]*object, error) {
		if start.dirEntry.IsDir() {
			if next != nil {
				chs, err := start.getChildren()
				if err != nil {
					return nil, err
				}
				res := []*object{}
				for _, ch := range chs {
					chRes, err := next(ch)
					if err != nil {
						return nil, err
					}
					res = append(res, chRes...)
				}
				return res, nil
			}
			return []*object{start}, nil
		}
		return []*object{}, nil
	}
}

type patternOp struct {
	rex *regexp.Regexp
}

func (opp *patternOp) opProduce(next operator) operator {
	return func(start *object) ([]*object, error) {
		if opp.rex.Match([]byte(start.dirEntry.Name())) {
			if next != nil {
				return next(start)
			}
			return []*object{start}, nil
		}
		return []*object{}, nil
	}
}

type bootOp struct{}

func (opp *bootOp) opProduce(next operator) operator {
	return func(start *object) ([]*object, error) {
		chs, err := start.getChildren()
		if err != nil {
			return nil, err
		}
		res := []*object{}
		for _, ch := range chs {
			chRes, err := next(ch)
			if err != nil {
				return nil, err
			}
			res = append(res, chRes...)
		}
		return res, nil
	}
}

type rootOp struct{}

func (opp *rootOp) opProduce(next operator) operator {
	return func(_ *object) ([]*object, error) {
		return next(site)
	}
}

type relativeOp struct {
	count int
}

func (opp *relativeOp) opProduce(next operator) operator {
	return func(start *object) ([]*object, error) {
		anch := start
		for i := 1; i < opp.count; i++ {
			anch = anch.parent
		}
		if next != nil {
			return next(anch)
		}
		return []*object{anch}, nil
	}
}

type recurseOp struct{}

func walk(start *object, next operator) ([]*object, error) {
	chs, err := start.getChildren()
	if err != nil {
		return nil, err
	}
	res := []*object{}
	for _, ch := range chs {
		if next != nil {
			chRes, err := next(ch)
			if err != nil {
				return nil, err
			}
			res = append(res, chRes...)
		} else {
			res = append(res, ch)
		}
		if ch.dirEntry.IsDir() {
			chRes, err := walk(ch, next)
			if err != nil {
				return nil, err
			}
			res = append(res, chRes...)
		}
	}
	return res, nil
}

func (opp *recurseOp) opProduce(next operator) operator {
	return func(start *object) ([]*object, error) {
		return walk(start, next)
	}
}
