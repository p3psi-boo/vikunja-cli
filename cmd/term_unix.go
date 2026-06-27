//go:build darwin || linux || freebsd || netbsd || openbsd

package cmd

import (
	"fmt"

	"golang.org/x/term"
)

// makeRawNoEcho puts the given terminal file descriptor into raw mode so that
// input (e.g. a password) is not echoed. It returns the previous state to be
// restored via restoreTerm. It returns an error when fd is not a terminal.
func makeRawNoEcho(fd int) (any, error) {
	return term.MakeRaw(fd)
}

// restoreTerm restores the terminal to its previous state.
func restoreTerm(fd int, state any) error {
	s, ok := state.(*term.State)
	if !ok || s == nil {
		return fmt.Errorf("invalid terminal state")
	}
	return term.Restore(fd, s)
}
