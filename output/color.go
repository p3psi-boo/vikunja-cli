package output

import (
	"io"
	"os"
	"strings"
	"sync"

	"github.com/fatih/color"
	"golang.org/x/term"
)

// ColorMode controls whether colorized output is emitted.
type ColorMode int

const (
	// ColorAuto enables color when the writer is an interactive terminal and
	// NO_COLOR is unset (the de-facto standard https://no-color.org).
	ColorAuto ColorMode = iota
	// ColorAlways forces color codes regardless of the destination.
	ColorAlways
	// ColorNever disables all color codes.
	ColorNever
)

var (
	colorMu      sync.Mutex
	colorMode    = ColorAuto
	colorNoColor = false // forced off via NO_COLOR
	// outputTarget is the writer auto-mode inspects; defaults to os.Stdout.
	outputTarget io.Writer = os.Stdout
)

// SetColorMode configures the global color policy. Safe to call at startup.
func SetColorMode(mode ColorMode) {
	colorMu.Lock()
	defer colorMu.Unlock()
	colorMode = mode
	applyColorMode()
}

// InitColorFromEnv records whether NO_COLOR is set. Must be called once at
// startup, before any output is produced. Honored by ColorAuto.
func InitColorFromEnv() {
	colorMu.Lock()
	defer colorMu.Unlock()
	colorNoColor = lookupNoColor()
	applyColorMode()
}

// SetColorTarget records the writer auto-mode inspects for terminal-ness.
// Intended for tests; in production stdout is used.
func SetColorTarget(w io.Writer) {
	colorMu.Lock()
	defer colorMu.Unlock()
	outputTarget = w
	applyColorMode()
}

func lookupNoColor() bool {
	// Per the spec, NO_COLOR need only be set (any value, including empty).
	_, ok := os.LookupEnv("NO_COLOR")
	return ok
}

func applyColorMode() {
	enabled := computeColorEnabled()
	color.NoColor = !enabled
}

func computeColorEnabled() bool {
	switch colorMode {
	case ColorAlways:
		return true
	case ColorNever:
		return false
	default: // ColorAuto
		if colorNoColor {
			return false
		}
		return isWriterTerminal(outputTarget)
	}
}

func isWriterTerminal(w io.Writer) bool {
	if w == nil {
		return false
	}
	if f, ok := w.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}

// ParseColorMode converts a user-supplied flag value into a ColorMode.
// Empty input yields ColorAuto.
func ParseColorMode(value string) (ColorMode, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "auto":
		return ColorAuto, true
	case "always", "force", "on", "yes":
		return ColorAlways, true
	case "never", "off", "none", "no":
		return ColorNever, true
	default:
		return ColorAuto, false
	}
}

// --- Semantic styling helpers ---------------------------------------------
//
// These wrap fatih/color attributes. They return the input unchanged when
// color is disabled (color.NoColor == true), so callers can pass strings
// through unconditionally and stay readable on pipes / redirected output.

// Overdue marks a value that is past its due date.
func Overdue(s string) string { return colorize(s, color.FgRed, color.Bold) }

// DueSoon marks a value due within the near horizon (today / tomorrow).
func DueSoon(s string) string { return colorize(s, color.FgYellow) }

// DoneStyle marks completed items in a muted, de-emphasized tone.
func DoneStyle(s string) string { return colorize(s, color.Faint) }

// PriorityStyle emphasizes priority values; higher numbers get stronger color.
func PriorityStyle(value int, s string) string {
	switch {
	case value >= 4:
		return colorize(s, color.FgRed, color.Bold)
	case value >= 2:
		return colorize(s, color.FgYellow)
	default:
		return s
	}
}

// Muted renders secondary information such as counts and hints.
func Muted(s string) string { return colorize(s, color.Faint) }

// Bold emphasizes headings and separators.
func Bold(s string) string { return colorize(s, color.Bold) }

// Label renders a label badge.
func Label(s string) string { return colorize(s, color.FgCyan) }

// FavoriteMark returns the symbol used to flag favorites.
func FavoriteMark() string { return "★" }

// CheckMark returns the symbol used for an affirmative boolean / done state.
func CheckMark() string { return "✓" }

// Empty returns the symbol used for an unset / negative value.
func EmptyMark() string { return "—" }

func colorize(s string, attrs ...color.Attribute) string {
	if color.NoColor {
		return s
	}
	c := color.New(attrs...)
	return c.Sprint(s)
}
