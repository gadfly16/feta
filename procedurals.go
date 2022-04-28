package feta

type objProc struct{}

var procedurals fDict = fDict{
	"obj": &objProc{},
}

func (node *objProc) eval(ctx *context) fExpr {
	fi, err := ctx.obj.dirEntry.Info()
	if err != nil {
		return fError{err.Error()}
	}
	return fDict{
		"name":  fString(fi.Name()),
		"isDir": fBool(fi.IsDir()),
		"size":  fNumber(fi.Size()),
	}
}
