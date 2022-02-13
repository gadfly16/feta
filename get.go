package feta

import (
	"encoding/json"
	"fmt"
)

func Get(query string, workDir string) ([]byte, error) {
	workDirObj, err := getObject(workDir)
	if err != nil {
		return nil, fmt.Errorf("Couldn't get object for workdir '%s': %v", workDir, err)
	}
	ast, err := Parse(query, []byte(query))
	if err != nil {
		return nil, fmt.Errorf("Couldn't parse query '%s': %v", query, err)
	}
	objs, err := ast.(operator)(workDirObj)
	if err != nil {
		return nil, fmt.Errorf("Couldn't evaluate query '%s': %v", query, err)
	}
	res, err := json.MarshalIndent(objs, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("Couldn't marshal objects: %v", err)
	}
	return res, nil
}
