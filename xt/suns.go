package xt

func GetSuns(xf Xfiles, text Text) []TSun {
	f := xf.Open("addon/types/TSuns.txt")
	defer f.Close()
	ret := []TSun{}
	tparse(f, text, &ret)
	return ret
}

// the only documentation of this I found was wrong.
type TSun struct {
	Unknown01  int
	Unknown02  int
	Unknown03  int
	Unknown04  int
	Unknown05  float64
	Unknown06  int
	Unknown07  int
	Unknown08  int
	Unknown09  int
	Brightness int // Seems to be fixed point scaled by 2^16
	Unknown11  int
	Unknown12  int
	Unknown13  int
	Unknown14  int
	Unknown15  int
	Unknown16  int
	Unknown17  int
	Unknown18  int
	ObjectID   string
}
