package feta

//go:generate pigeon -o parser.go parser.peg

import "regexp"

type oNode interface {
	get(*object) ([]*object, error)
}

type levelNode struct {
	directory bool
	name      *nameMatch
	next      *levelNode
}

type nameMatch struct {
	rex *regexp.Regexp
}

func (node *levelNode) get(start *object) ([]*object, error) {
	chs, err := start.getChildren()
	if err != nil {
		return nil, err
	}
	res := []*object{}
	for _, ch := range chs {
		if node.name.match(ch) {
			chIsDir := ch.dirEntry.IsDir()
			if node.next != nil {
				if chIsDir {
					chRes, err := node.next.get(ch)
					if err != nil {
						return nil, err
					}
					res = append(res, chRes...)
				}
				continue
			}
			if node.directory {
				if !chIsDir {
					continue
				}
			}
			res = append(res, ch)
		}
	}
	return res, nil
}

type patternNode struct {
	pattern string
}

func (node *nameMatch) match(obj *object) bool {
	return node.rex.Match([]byte(obj.dirEntry.Name()))
}
