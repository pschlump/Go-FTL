//
// Go-FTL / TabServer2
//
// Copyright (C) Philip Schlump, 2012-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1011
//

package TabServer2

import "fmt"

type LoginSystemType int

const (
	LstNone   LoginSystemType = 1
	LstAesSrp LoginSystemType = 2
	LstUnPw   LoginSystemType = 3
	LstBasic  LoginSystemType = 4
)

func (lst LoginSystemType) String() string {
	switch lst {
	case LstNone:
		return "LstNone"
	case LstAesSrp:
		return "LstAesSrp"
	case LstUnPw:
		return "LstUnPw"
	case LstBasic:
		return "LstBasic"
	default:
		return fmt.Sprintf("--- Unknown LoginSystemType=%d ---", int(lst))
	}
}
