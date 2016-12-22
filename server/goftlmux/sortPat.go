package goftlmux

//
// Go Go Mux - Go Fast Mux / Router for HTTP requests
//
// (C) Philip Schlump, 2013-2014.
// Version: 0.4.3
// BuildNo: 804
//
//

// Max Slash needs to be a var - we go above we - don't need to search through or anything

// xyzzy202 - need a test - need to do someting with {} patterns - need to deal with {name} patterns

import "sort"

// -------------------------------------------------------------------------------------------------
// Sorting of patterns.   Use to improve the average case - at the expense of more rare cases.
// -------------------------------------------------------------------------------------------------

type lessFuncPat func(p1, p2 *UrlAPat) bool

// multiSorter implements the Sort interface, sorting the the_data within.
type multiSorterPat struct {
	the_data []UrlAPat
	less     []lessFuncPat
}

// Sort sorts the argument slice according to the less functions passed to OrderedBy.
// var nMatch []UrlPat  // Index by Length ( NSl )
func (ms *multiSorterPat) Sort(the_data []UrlAPat) {
	ms.the_data = the_data
	sort.Sort(ms)
}

// OrderedBy returns a Sorter that sorts using the less functions, in order.
// Call its Sort method to sort the data.
func OrderedByPat(less ...lessFuncPat) *multiSorterPat {
	return &multiSorterPat{
		less: less,
	}
}

// Len is part of sort.Interface.
func (ms *multiSorterPat) Len() int {
	return len(ms.the_data)
}

// Swap is part of sort.Interface.
func (ms *multiSorterPat) Swap(i, j int) {
	ms.the_data[i], ms.the_data[j] = ms.the_data[j], ms.the_data[i]
}

// Less is part of sort.Interface. It is implemented by looping along the
// less functions until it finds a comparison that is either Less or
// !Less. Note that it can call the less functions twice per call. We
// could change the functions to return -1, 0, 1 and reduce the
// number of calls for greater efficiency: an exercise for the reader.
func (ms *multiSorterPat) Less(i, j int) bool {
	p, q := &ms.the_data[i], &ms.the_data[j]
	// Try all but the last comparison.
	var k int
	for k = 0; k < len(ms.less)-1; k++ {
		less := ms.less[k]
		switch {
		case less(p, q):
			// p < q, so we have a decision.
			return true
		case less(q, p):
			// p > q, so we have a decision.
			return false
		}
		// p == q; try the next comparison.
	}
	// All comparisons to here said "equal", so just return whatever
	// the final comparison reports.
	return ms.less[k](p, q)
}
