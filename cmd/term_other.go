//go:build !(darwin || linux || freebsd || netbsd || openbsd)

package cmd

import "fmt"

// makeRawNoEcho is unsupported on this platform; promptHidden falls back to a
// normal prompt.
func makeRawNoEcho(fd int) (any, error) {
	return nil, fmt.Errorf("not a supported terminal")
}

func restoreTerm(fd int, state any) error {
	return nil
}
