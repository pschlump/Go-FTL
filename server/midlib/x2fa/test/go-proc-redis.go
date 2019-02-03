package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// writeLines writes the lines to the given file.
func writeLines(lines []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

func main() {
	lines, err := readLines(",r1")
	if err != nil {
		log.Fatalf("readLines: %s", err)
	}
	// for i, line := range lines {
	for i := 1; i < len(lines); i++ {
		// fmt.Println(i, line)
		fmt.Printf("get %s\nttl %s\n\n", lines[i], lines[i])
	}

	//	if err := writeLines(lines, "foo.out.txt"); err != nil {
	//		log.Fatalf("writeLines: %s", err)
	//	}
}
