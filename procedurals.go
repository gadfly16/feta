package feta

type objProc struct{}

var procedurals fDict = fDict{
	"obj": &objProc{},
}

func (node *objProc) eval(ctx *context) (fType, error) {
	fi, err := ctx.obj.dirEntry.Info()
	if err != nil {
		return nil, err
	}
	return fDict{
		"name":  fString(fi.Name()),
		"isDir": fBool(fi.IsDir()),
		"size":  fNumber(fi.Size()),
	}, nil
}
