package feta

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	site *object
)

type object struct {
	dirEntry  os.DirEntry
	parent    *object
	isProjSet bool
	project   *object
	children  []*object
	meta      fDict
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

func (o *object) sysPath() string {
	path := ""
	if o.dirEntry.IsDir() {
		path = "/"
	}
	for o != site {
		path = "/" + o.dirEntry.Name() + path
		o = o.parent
	}
	if Flags.SitePath != "/" {
		path = Flags.SitePath + path
	}
	return path
}

func (o *object) fetaPath() string {
	if o == site {
		return "/"
	}
	path := ""
	if o.dirEntry.IsDir() {
		path = "/"
	}
	for o != site {
		path = "/" + o.dirEntry.Name() + path
		o = o.parent
	}
	return path
}

func (o *object) MarshalJSON() ([]byte, error) {
	if Flags.SysAbs {
		return json.Marshal(o.sysPath())
	}
	return json.Marshal(o.fetaPath())
}

func (o *object) getChildren() ([]*object, error) {
	if o.children == nil {
		des, err := os.ReadDir(o.sysPath())
		if err != nil {
			return nil, err
		}
		for _, de := range des {
			if de.Name() != ".feta" {
				o.children = append(o.children, newObject(o, de))
			}
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
	if path == Flags.SitePath {
		return site, nil
	}
	o, err := site.find(strings.Split(trimSitePath(path), "/"))
	return o, err
}

func trimSitePath(path string) string {
	if Flags.SitePath == "/" {
		return strings.TrimPrefix(path, Flags.SitePath)
	}
	return strings.TrimPrefix(path, Flags.SitePath+"/")
}

func InitSite(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("Couldn't create absolute path from '%s': %v", path, err)
	}
	fi, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("Couldn't stat site path '%s': %v", absPath, err)
	}
	site = newObject(nil, fs.FileInfoToDirEntry(fi))
	Log(fmt.Sprintf("Site set to: %s", absPath))
	return absPath, nil
}

func (o *object) getMeta() (fDict, error) {
	if o.meta != nil {
		return o.meta, nil
	}
	path := o.sysPath()
	if o.dirEntry.IsDir() {
		path = path + ".feta/_"
	} else {
		path = filepath.Dir(path) + "/.feta/" + o.dirEntry.Name() + "._"
	}

	js, err := ioutil.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			meta := fDict{}
			insertProcedurals(meta)
			return meta, nil
		}
		return nil, fmt.Errorf("Couldn't read meta file: %v", err)
	}
	raw := make(map[string]interface{})
	err = json.Unmarshal(js, &raw)
	if err != nil {
		return nil, fmt.Errorf("Couldn't unmarshal meta file '%s': %v", path, err)
	}
	meta, err := typeConvert(raw)
	if err != nil {
		return nil, fmt.Errorf("Couldn't type-convert meta file '%s': %v", path, err)
	}
	insertProcedurals(meta.(fDict))
	o.meta = meta.(fDict)
	return o.meta, nil
}

func insertProcedurals(meta fDict) {
	for k, v := range procedurals {
		meta[k] = v
	}
}

func fileExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}

func (o *object) getProject() (*object, error) {
	if o.isProjSet {
		return o.project, nil
	}
	var proj *object
	var err error
	if o.dirEntry.IsDir() {
		path := o.sysPath() + "/.feta/project"
		if isProj, err := fileExists(path); err != nil {
			return nil, err
		} else if isProj {
			proj = o
		} else if o == site {
			proj = nil
		} else {
			proj, err = o.parent.getProject()
			if err != nil {
				return nil, err
			}
		}
	} else {
		proj, err = o.parent.getProject()
		if err != nil {
			return nil, err
		}
	}
	o.project = proj
	o.isProjSet = true
	return proj, nil
}

func (value *object) eval(ctx *context) fExpr {
	return value
}

func (value *object) boolVal() fBool {
	return true
}
