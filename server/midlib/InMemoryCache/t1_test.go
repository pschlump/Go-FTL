//
// Go-FTL
//
// Copyright (C) Philip Schlump, 2014-2016
//
// Do not remove the following lines - used in auto-update.
// Version: 0.5.9
// BuildNo: 1811
// FileId: 1253
//

package InMemoryCache

import "testing"

// -----------------------------------------------------------------------------------------------------------------------------------------------
// test components that don't require a server
//

func Test_Componetnts1(t *testing.T) {

	tests := []struct {
		num string
		cnv int64
		ok  bool
	}{
		{"100", 100, true},
		{"100M", 100 * 1024 * 1024, true},
		{"100Z", 0, false},
		{"100G", 100 * 1024 * 1024 * 1024, true},
		{"100T", 100 * 1024 * 1024 * 1024 * 1024, true},
		{"100g", 100 * 1000 * 1000 * 1000, true},
	}

	// fmt.Printf("Expect Output\nInvalid size input >100Z< expected 0..n, (will use all available space)\n---------------------------------------\n")

	for ii, test := range tests {
		_, _ = ii, test

		u, err := ConvertMGTPToValue(test.num)

		if err != nil {
			if test.ok == true {
				t.Errorf("Error %2d, returned error when expecting success\n", ii)
			}
		} else {
			if test.ok == false {
				t.Errorf("Error %2d, returned success wwhen expecting error\n", ii)
			}

		}

		if u != test.cnv {
			t.Errorf("Error %2d, got: %d, expected %d\n", ii, u, test.cnv)
		}

	}

}

/* vim: set noai ts=4 sw=4: */
