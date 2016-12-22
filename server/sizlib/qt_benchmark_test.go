package sizlib

import (
	"fmt"
	"testing"
)

func BenchmarkQt(b *testing.B) {
	var str string
	mm := make(map[string]string)
	mm["m1"] = "M1"

	for n := 0; n < b.N; n++ {
		str = Qt("m1=%{m1%} m2=%{m2%}", mm)
	}

	if ShowBenchResults { // gurantee that the compile will not optimize out the work
		fmt.Printf("%s\n", str)
	}
	//	b.StopTimer()
	//
	//	if s := strings.Repeat("x", b.N); str != s {
	//		b.Errorf("unexpected result; got=%s, want=%s", str, s)
	//	}
}

var ShowBenchResults = false
