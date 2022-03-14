package feta

type flags struct {
	Verbose  bool
	SitePath string
	SysAbs   bool
	UglyJSON bool
}

var Flags flags
