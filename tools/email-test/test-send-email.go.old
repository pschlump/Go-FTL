package main

import (
	"flag"

	aws "github.com/pschlump/Go-FTL/tools/email-test/lib"
)

// xyzzy - To, default to me
// xyzzy - see "make test8" in ../../ses-aws/email-cli

var UniqueSubject = flag.String("us", "", "A unique subject") // 0
func init() {
	flag.StringVar(UniqueSubject, "u", "", "A unique subject") // 0
}

func main() {

	var err error

	flag.Parse()

	aws.ConfigEmailAWS("./email-config.json")

	s1 := *UniqueSubject
	b1 := "HTML Body"
	b2 := "Text Body"

	err = aws.SendEmailViaAWS(s1, b1, b2, "pschlump@gmail.com")

	_ = err

}
