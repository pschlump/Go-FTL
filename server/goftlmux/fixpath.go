package goftlmux

import (
	debug "github.com/pschlump/godebug"
)

//
// Go Go Mux - Go Fast Mux / Router for HTTP requests
//
// (C) Philip Schlump, 2013-2014.
// Version: 0.4.3
// BuildNo: 804
//
//

// Parse a path and make corrections to it.   Remove double slash, "//".
// Remove all "./".   Convert "name/../" to empty string.  Return the
// number of used elements in the 'rv' slice, and place one token
// from the path in each of the elements of the 'rv' slice.  If the
// path is a hard path (Starts from /) then the 0th element of the
// rv slice will be "/".  The rv slice is assumed to be long enough
// to hold all the elements of the path.  max is the maximum number
// of elements that will be used in 'rv'.
func FixPath(pth string, rv []string, max int) int {
	l := len(pth)
	i := 0
	rv_n := 0
	beg := 0
	end := 0
	// rv_out := ""

	// if debug {
	// 	fmt.Printf("\nTest ->%s<-\n", pth)
	// }

s0:
	// if debug {
	// 	fmt.Printf("s0:   l=%d i=%d rv_n=%d beg=%d end=%d ->%s<-, rv=%s\n", l, i, rv_n, beg, end, pth, rvAsStr(rv, rv_n))
	// }
	if i >= l {
		goto s21
	} else if pth[i] == '/' {
		i++
		goto s1
	} else if pth[i] == '.' {
		beg = i
		i++
		goto s3
	} else {
		beg = i
		goto s10
	}
s1:
	//if debug {
	// fmt.Printf("s1:   l=%d i=%d rv_n=%d beg=%d end=%d ->%s<-, rv=%s\n", l, i, rv_n, beg, end, pth, rvAsStr(rv, rv_n))
	//}
	rv[rv_n] = "/"
	rv_n++
	if rv_n >= max {
		return rv_n
	}
s2:
	// if debug {
	// 	fmt.Printf("s2:   l=%d i=%d rv_n=%d beg=%d end=%d ->%s<-, rv=%s\n", l, i, rv_n, beg, end, pth, rvAsStr(rv, rv_n))
	// }
	if i >= l {
		goto s20
	} else if pth[i] == '/' {
		i++
		goto s2
	} else if pth[i] == '.' {
		beg = i
		i++
		goto s3
	} else {
		beg = i
		goto s10
	}
s3:
	// if debug {
	// 	fmt.Printf("s3:   l=%d i=%d rv_n=%d beg=%d end=%d ->%s<-, rv=%s\n", l, i, rv_n, beg, end, pth, rvAsStr(rv, rv_n))
	// }
	if i >= l {
		goto s20
	} else if pth[i] == '.' {
		i++
		goto s4
	} else if pth[i] == '/' {
		i++
		goto s2
	} else {
		goto s10
	}
s4:
	// if debug {
	// 	fmt.Printf("s4:   l=%d i=%d rv_n=%d beg=%d end=%d ->%s<-, rv=%s\n", l, i, rv_n, beg, end, pth, rvAsStr(rv, rv_n))
	// }
	if i >= l {
		goto s20
	} else if pth[i] == '/' {
		if rv_n > 0 {
			rv_n--
		}
		// if lenrv == 0 {
		if rv_n == 0 {
			goto s0
		}
		i++
		goto s2
	} else {
		goto s10
	}

s10:
	// if debug {
	// 	fmt.Printf("s10:   l=%d i=%d rv_n=%d beg=%d end=%d ->%s<-, rv=%s\n", l, i, rv_n, beg, end, pth, rvAsStr(rv, rv_n))
	// }
	if i >= l {
		end = i
		goto s15
	} else if pth[i] == '/' {
		end = i
		goto s15
	} else {
		i++
		goto s10
	}

s15:
	// if debug {
	// 	rv_out = "-- no save --"
	// 	if end > beg {
	// 		rv_out = pth[beg:end]
	// 	}
	// 	fmt.Printf("s15:   l=%d i=%d rv_n=%d beg=%d end=%d ->%s<-, rv=%s -- save ->%s<-\n", l, i, rv_n, beg, end, pth, rvAsStr(rv, rv_n), rv_out)
	// }
	// fmt.Printf("beg=%d end=%d, rv_n=%d, len(pth) = %d rv=%x\n", beg, end, rv_n, len(pth), rv)
	if end > beg {

		rv[rv_n] = pth[beg:end]
		rv_n++
		if rv_n >= max {
			return rv_n
		}
	}
	end = -1
	if i >= l {
		goto s20
	} else {
		goto s2
	}

s20:
	// if debug {
	// 	fmt.Printf("s20/Final  l=%d i=%d rv_n=%d beg=%d end=%d ->%s<-, rv=%s\n", l, i, rv_n, beg, end, pth, rvAsStr(rv, rv_n))
	// }
	rv = rv[:rv_n]
	return rv_n

s21:
	// if debug {
	// 	fmt.Printf("s21/Top  l=%d i=%d rv_n=%d beg=%d end=%d ->%s<-, rv=%s\n", l, i, rv_n, beg, end, pth, rvAsStr(rv, rv_n))
	// }
	rv[rv_n] = "/"
	// if debug {
	// 	fmt.Printf("s21/Final  l=%d i=%d rv_n=%d beg=%d end=%d ->%s<-, rv=%s\n", l, i, rv_n, beg, end, pth, rvAsStr(rv, rv_n))
	// }
	rv = rv[:rv_n+1]
	return rv_n + 1
}

// var debug = false

func rvAsStr(rv []string, n int) (s string) {
	s = debug.SVar(rv[0:n])
	return
}
