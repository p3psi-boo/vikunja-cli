//go:build !(darwin || linux || freebsd || netbsd || openbsd)

package cmd

import "io"

// readPassword is unsupported on this platform; the caller falls back to an
// echoed prompt.
func readPassword(fd int, out io.Writer, label string) (value string, ok bool, err error) {
	return "", false, nil
}
