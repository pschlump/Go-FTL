package Acb1

import "os"

var hasBeenSetup bool = false

func (hdlr *Acb1Type) SetupServer() {
	// TODO - should check path to see if it exits
	if !hasBeenSetup {
		os.Mkdir(hdlr.OutputPath, 0755)
		hasBeenSetup = true
	}
}
