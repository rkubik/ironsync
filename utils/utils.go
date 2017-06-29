package utils

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

var (
	chunkSize     = 64000
	unixEpochTime = time.Unix(0, 0)
)

// DeepCompare - Compare two files to see if they contain the same
// content. If an errors occur false is returned.
// Adapted from:
// https://stackoverflow.com/questions/29505089/how-can-i-compare-two-files-in-golang/30038571#30038571
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

// IsZeroTime reports whether t is obviously unspecified (either zero or Unix()=0).
func IsZeroTime(t time.Time) bool {
	return t.IsZero() || t.Equal(unixEpochTime)
}

// PublicKeyFile reads public key's from a private key file into memory
func PublicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}

	return ssh.PublicKeys(key)
}

// RunCmd executes a cmd from a string (todo: better support)
func RunCmd(cmd string, timeout int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	args := strings.Split(cmd, " ")
	err := exec.CommandContext(ctx, args[0], args[1:]...).Run()
	return err
}
