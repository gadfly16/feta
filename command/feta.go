package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/tidwall/pretty"

	"github.com/gadfly16/feta"
)

func defineFlags(homeDir string) {
	if flag.Lookup("v") == nil {
		flag.BoolVar(&feta.Flags.Verbose, "v", false, "Verbose output")
		flag.StringVar(&feta.Flags.SitePath, "S", homeDir, "Site directory path")
		flag.BoolVar(&feta.Flags.SysAbs, "a", false, "System absolute output")
		flag.BoolVar(&feta.Flags.UglyJSON, "u", false, "Ugly JSON output")
	}
}

var out io.Writer = os.Stdout

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		feta.Fatal(err)
	}

	defineFlags(homeDir)
	flag.Parse()

	absSitePath, err := feta.InitSite(feta.Flags.SitePath)
	if err != nil {
		feta.Fatal(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		feta.Fatal(err)
	}

	if !strings.HasPrefix(wd, absSitePath) {
		feta.Fatal("Invocation dir must be under site path:" + absSitePath)
	}

	switch flag.Arg(0) {
	case "get":
		res, err := feta.Get(flag.Arg(1), wd)
		if err != nil {
			feta.Fatal(err)
		}
		if feta.Flags.UglyJSON {
			res = append(res, '\n')
		} else {
			res = pretty.Color(pretty.Pretty(res), nil)
		}
		fmt.Fprint(out, string(res))
	// case "set":
	// 	setCmd()
	default:
		feta.Fatal("Unknown command: " + flag.Arg(0))
	}
}
