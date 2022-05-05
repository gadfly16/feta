package feta

import (
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
	res := ast.(fExpr).eval(&context{obj: workDirObj})
	return marshal(res.(fNode), !Flags.UglyJSON), nil
}
