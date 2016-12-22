package testsup

import "github.com/pschlump/Go-FTL/server/goftlmux"

type NameValue struct {
	Name  string
	Value string
}

func SetupTestCreateHeaders(wr *goftlmux.MidBuffer, hdr []NameValue) {
	for _, hh := range hdr {
		wr.Headers.Set(hh.Name, hh.Value)
	}
}
