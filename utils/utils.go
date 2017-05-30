package utils

import (
	"bytes"
	"io"
	"os"
)

const chunkSize = 64000

// DeepCompare - Compare two files to see if they contain the same
// content. If an errors occur false is returned.
func DeepCompare(file1, file2 string) (equal bool) {
	f1, err1 := os.Open(file1)
	if err1 != nil {
		return
	}
	defer f1.Close()

	f2, err2 := os.Open(file2)
	if err2 != nil {
		return
	}
	defer f2.Close()

	for {
		b1 := make([]byte, chunkSize)
		_, cmpErr1 := f1.Read(b1)

		b2 := make([]byte, chunkSize)
		_, cmpErr2 := f2.Read(b2)

		if cmpErr1 != nil || cmpErr2 != nil {
			if cmpErr1 == io.EOF && cmpErr2 == io.EOF {
				return true
			} else if cmpErr1 == io.EOF || cmpErr2 == io.EOF {
				return false
			} else {
				return false
			}
		}

		if !bytes.Equal(b1, b2) {
			return false
		}
	}
}
