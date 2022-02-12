package feta

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var (
	root     *object
	rootPath string
)

type object struct {
	dirEntry os.DirEntry
	parent   *object
	children []*object
	meta     MMap
}

func newObject(parent *object, dirEntry os.DirEntry) (o *object) {
	o = &object{
		dirEntry: dirEntry,
		parent:   parent,
		children: nil,
		meta:     nil,
	}
	return o
}

func (o *object) path() string {
	if o == root {
		return rootPath
	}
	path := ""
	for o != root {
		path = "/" + o.dirEntry.Name() + path
		o = o.parent
	}
	if rootPath != "/" {
		path = rootPath + path
	}
	return path
}

func (o *object) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.path())
}

func (o *object) getChildren() ([]*object, error) {
	if o.children == nil {
		des, err := os.ReadDir(o.path())
		if err != nil {
			return nil, err
		}
		for _, de := range des {
			o.children = append(o.children, newObject(o, de))
		}
	}
	return o.children, nil
}

func (o *object) find(pathList []string) (*object, error) {
	chs, err := o.getChildren()
	if err != nil {
		return nil, err
	}
	for _, ch := range chs {
		if pathList[0] == ch.dirEntry.Name() {
			if len(pathList) == 1 {
				return ch, nil
			}
			return ch.find(pathList[1:])
		}
	}
	return nil, errors.New("Can't find it..")
}

func getObject(path string) (*object, error) {
	if path == rootPath {
		return root, nil
	}
	o, err := root.find(strings.Split(trimRootPath(path), "/"))
	return o, err
}

func trimRootPath(path string) string {
	if rootPath == "/" {
		return strings.TrimPrefix(path, rootPath)
	}
	return strings.TrimPrefix(path, rootPath+"/")
}

func InitRoot(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("Couldn't create absolute path from '%s': %v", path, err)
	}
	rootPath = absPath
	fi, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("Couldn't init root '%s': %v", absPath, err)
	}
	root = newObject(nil, fs.FileInfoToDirEntry(fi))
	Log(fmt.Sprintf("Root set to: %s", absPath))
	return nil
}
