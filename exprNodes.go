package feta

import "errors"

type exprNode interface {
	eval(fType) (fType, error)
}

type keyNode struct {
	key  string
	next exprNode
}

func (node *keyNode) eval(val fType) (fType, error) {
	switch t := val.(type) {
	case fDict:
		res, exists := t[node.key]
		if exists {
			if node.next == nil {
				return res, nil
			}
			return node.next.eval(res)
		}
		return nil, nil
	}
	return nil, errors.New("Trying to access name in non-object type.")
}
