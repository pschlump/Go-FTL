package goftlmux

//
// Go Go Mux - Go Fast Mux / Router for HTTP requests
//
// (C) Philip Schlump, 2013-2014.
// Version: 0.4.3
// BuildNo: 804
//
//

// -------------------------------------------------------------------------------------------------
/* USED */
// Test: parseReFromToken_test.go

// Possible Improvement - should we validate that name is a name - [a-zA-Z_][a-zA-Z_0-9]* - Maybee?

// Pars	{name:Re} into the name and the regular expression.  Indicate with convertToColon==true
// that this is a {name} pattern that matches /[^/]*/
func parseReFromToken3(s string) (name string, re string, valid bool, convertToColon bool) {
	var i = 0
	var sp = 0
	name = ""
	re = ""
	valid = false
	convertToColon = false

	if len(s) <= 0 {
		return
	}

	// fmt.Printf("i=%d len(s)=%d\n", i, len(s))
	if i < len(s) && s[i] == '{' {
		i++
	}
	sp = i

	for i < len(s) && s[i] != ':' && s[i] != '}' {
		i++
	}
	if i >= len(s) {
		return
	}
	if i < len(s) {
		name = s[sp:i]
	}
	if s[i] == '}' {
		if len(name) == 0 {
			return
		}
		convertToColon = true
		valid = true
		return
	}

	if i < len(s) && s[i] == ':' {
		i++
	}
	sp = i
	var d = 0
	for i < len(s) {
		if s[i] == '{' {
			d++
		} else if s[i] == '}' {
			if d == 0 {
				break
			}
			d--
		}
		i++
	}
	if i < len(s) {
		re = s[sp:i]
	}

	if len(name) > 0 && len(re) > 0 {
		valid = true
	}
	return
}
