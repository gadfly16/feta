package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/gadfly16/feta"
)

func defineFlags(homeDir string) {
	flag.BoolVar(&feta.Flags.Verbose, "v", false, "Verbose output")
	flag.StringVar(&feta.Flags.SitePath, "S", homeDir, "Site directory path")
	flag.BoolVar(&feta.Flags.SysAbs, "a", false, "System absolute output")
}

func setCmd() {
}

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
		fmt.Println(string(res))
	case "set":
		setCmd()
	default:
		feta.Fatal("Unknown command: " + flag.Arg(0))
	}

	// m := make(ul.MMap)
	// m["number"] = ul.MNumber(123.34)
	// m["text"] = ul.MText("breeze")
	// m["time"] = ul.MTime(time.Now())
	// m["expression"] = ul.MExp("1+1")

	// ul.Log(m)
	// fmt.Printf("%T\n", m["number"])
	// j, err := json.Marshal(m)
	// if err != nil {
	// 	ul.Fatal(err)
	// }
	// fmt.Printf("%s\n", j)

	// m_ := make(map[string]interface{})
	// json.Unmarshal(j, &m_)
	// fmt.Println(m_)
	// m__, err := ul.TypeConvert(m_)
	// if err != nil {
	// 	ul.Fatal(err)
	// }
	// switch v := m__.(type) {
	// case ul.MMap:
	// 	fmt.Printf("%T\n", v["time"])
	// }
}
