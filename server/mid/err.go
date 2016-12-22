//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1197
//

package mid

import "errors"

var FtlConfigError = errors.New("Invalid type supplied to configuration function")
var ErrNonMidBufferWriter = errors.New("Invalid type - needs to be a goftlmux.MidBuffer")
var ErrInvalidConfiguration = errors.New("Invalid Configuration")
var ErrInternalError = errors.New("Internal Error")
var ErrMuxError = errors.New("Mux reported an error")
