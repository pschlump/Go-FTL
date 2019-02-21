package main

import (
	"fmt"
	"path/filepath"
)

func main() {
	TemplateFn := filepath.Clean("/.././a.b")
	fmt.Printf("output ->%s<-\n", TemplateFn)
}
