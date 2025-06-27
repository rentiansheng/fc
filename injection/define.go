package injection

/***************************
    @author: tiansheng.ren
    @date: 2024/12/19
    @desc:

***************************/

type Parser interface {
	Inject() (map[string][]*Inject, error)
	AddParesFile(file string)
}

var parser Parser

type Inject struct {
	St   uint32 `json:"struct"`
	Meta map[string][]*Inject
}
