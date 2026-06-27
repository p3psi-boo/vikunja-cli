//go:build darwin || linux || freebsd || netbsd || openbsd

package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// readPassword prompts for a line of input on fd and echoes an asterisk per
// typed character (with backspace support), returning the trimmed value. It
// puts the terminal into raw mode and restores it afterwards via defer, so the
// tty is never left in raw mode even on early return or interrupt.
//
// ok is true when the fd was a terminal the call knew how to handle; when false
// the caller should fall back to an echoed prompt.
//
// We read straight from the fd rather than the caller's bufio.Reader: that
// reader may have buffered bytes already, and switching the tty to raw mode
// underneath it leaves those bytes stranded. We also do the echo ourselves
// (MakeRaw disables kernel echo), terminating on either '\r' or '\n' since raw
// mode no longer turns the Return key into '\n'.
func readPassword(fd int, out io.Writer, label string) (value string, ok bool, err error) {
	if !term.IsTerminal(fd) {
		return "", false, nil
	}

	if _, err := fmt.Fprintf(out, "%s: ", label); err != nil {
		return "", true, err
	}

	old, err := term.MakeRaw(fd)
	if err != nil {
		return "", true, err
	}
	// Restore the terminal no matter how we leave, including interrupts.
	// Errors from Restore are secondary to the read result; report the first
	// real error below.
	defer func() { _ = term.Restore(fd, old) }()

	var buf []byte
	one := make([]byte, 1)
	for {
		n, readErr := os.NewFile(uintptr(fd), "stdin").Read(one)
		if n > 0 {
			switch c := one[0]; c {
			case '\r', '\n':
				fmt.Fprintln(out)
				return strings.TrimSpace(string(buf)), true, readErr
			case 0x7f, '\b': // DEL / Backspace
				if len(buf) > 0 {
					buf = buf[:len(buf)-1]
					fmt.Fprint(out, "\b \b") // erase the last '*'
				}
				continue
			case 0x03: // Ctrl-C
				fmt.Fprintln(out, "^C")
				return "", true, fmt.Errorf("interrupted")
			case 0x04: // Ctrl-D
				fmt.Fprintln(out)
				if len(buf) == 0 {
					return "", true, io.EOF
				}
				return strings.TrimSpace(string(buf)), true, readErr
			default:
				if c >= 0x20 && c < 0x7f { // printable ASCII
					buf = append(buf, c)
					fmt.Fprint(out, "*")
				}
				// Non-printable bytes are dropped silently.
			}
		}
		if readErr != nil {
			return strings.TrimSpace(string(buf)), true, readErr
		}
	}
}
