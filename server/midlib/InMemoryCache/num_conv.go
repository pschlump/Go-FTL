package InMemoryCache

import (
	"errors"
	"regexp"
	"strconv"
)

var numRe *regexp.Regexp
var mulFact map[string]int64

func init() {
	numRe = regexp.MustCompile(`^([0-9]+)([MGTPkmgtpk])?$`)
	mulFact = make(map[string]int64)
	mulFact[""] = 1
	mulFact["K"] = 1024
	mulFact["M"] = 1024 * 1024
	mulFact["G"] = 1024 * 1024 * 1024
	mulFact["T"] = 1024 * 1024 * 1024 * 1024
	mulFact["P"] = 1024 * 1024 * 1024 * 1024 * 1024
	mulFact["k"] = 1000
	mulFact["m"] = 1000 * 1000
	mulFact["g"] = 1000 * 1000 * 1000
	mulFact["t"] = 1000 * 1000 * 1000 * 1000
	mulFact["p"] = 1000 * 1000 * 1000 * 1000 * 1000
}

var ErrInvalidConversion = errors.New("Invalid number converstion")

func ConvertMGTPToValue(in string) (u int64, err error) {

	// extract number and MGTP, multiply for MGTP
	a := numRe.FindStringSubmatch(in)
	// fmt.Printf("db: len(a) = %d, a=%v\n", len(a), a)
	if len(a) == 2 || len(a) == 3 {
		u, err = strconv.ParseInt(a[1], 10, 64)
		if err != nil {
			// fmt.Printf("Invalid size input >%s< expected number, %s (will use all available space)\n", in, err)
			err = ErrInvalidConversion
			u = 0
		}
	}
	if len(a) == 3 {
		if f, ok := mulFact[a[2]]; !ok {
			// fmt.Printf("Invalid size input >%s< expected M, G, T, P (will use all available space)\n", in)
			err = ErrInvalidConversion
			u = 0
		} else {
			u *= f
		}
	}
	if len(a) <= 1 || len(a) > 3 || u < 0 {
		// fmt.Printf("Invalid size input >%s< expected 0..n, (will use all available space)\n", in)
		err = ErrInvalidConversion
		u = 0
	}
	return // 0 implies size based on available size
}

/* vim: set noai ts=4 sw=4: */
