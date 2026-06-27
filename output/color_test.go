package output

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/fatih/color"
)

func TestParseColorMode(t *testing.T) {
	cases := []struct {
		in      string
		want    ColorMode
		wantOK  bool
	}{
		{"", ColorAuto, true},
		{"auto", ColorAuto, true},
		{"AUTO", ColorAuto, true},
		{"always", ColorAlways, true},
		{"force", ColorAlways, true},
		{"never", ColorNever, true},
		{"off", ColorNever, true},
		{"banana", ColorAuto, false},
	}
	for _, tc := range cases {
		got, ok := ParseColorMode(tc.in)
		if got != tc.want || ok != tc.wantOK {
			t.Errorf("ParseColorMode(%q) = %v,%v; want %v,%v", tc.in, got, ok, tc.want, tc.wantOK)
		}
	}
}

func TestColorNeverStripsEscape(t *testing.T) {
	defer restoreColorDefaults()
	SetColorMode(ColorNever)
	out := Overdue("late")
	if strings.Contains(out, "\x1b") {
		t.Fatalf("ColorNever must not emit escape codes, got %q", out)
	}
	if out != "late" {
		t.Fatalf("ColorNever should pass string through, got %q", out)
	}
}

func TestColorAlwaysEmitsEscape(t *testing.T) {
	defer restoreColorDefaults()
	SetColorMode(ColorAlways)
	out := Overdue("late")
	if !strings.Contains(out, "\x1b") {
		t.Fatalf("ColorAlways must emit escape codes, got %q", out)
	}
	if !strings.Contains(out, "late") {
		t.Fatalf("output should contain the original string, got %q", out)
	}
}

func TestColorAutoRespectsNoColor(t *testing.T) {
	defer restoreColorDefaults()
	t.Setenv("NO_COLOR", "1")
	InitColorFromEnv()
	// non-tty buffer => auto should be disabled regardless
	SetColorTarget(&bytes.Buffer{})
	out := DueSoon("soon")
	if strings.Contains(out, "\x1b") {
		t.Fatalf("auto with NO_COLOR/non-tty must not emit escape codes, got %q", out)
	}
}

func TestColorAutoStripsEscapeOnNonTty(t *testing.T) {
	defer restoreColorDefaults()
	os.Unsetenv("NO_COLOR")
	colorNoColor = false
	SetColorTarget(&bytes.Buffer{}) // not a terminal
	SetColorMode(ColorAuto)
	out := Bold("hi")
	if strings.Contains(out, "\x1b") {
		t.Fatalf("auto on non-tty must not emit escape codes, got %q", out)
	}
}

func TestPriorityStyleHighValueColored(t *testing.T) {
	defer restoreColorDefaults()
	SetColorMode(ColorAlways)
	if out := PriorityStyle(5, "5"); !strings.Contains(out, "\x1b") {
		t.Fatalf("priority 5 should be colored, got %q", out)
	}
}

func TestPriorityStyleZeroPlain(t *testing.T) {
	defer restoreColorDefaults()
	SetColorMode(ColorAlways)
	if out := PriorityStyle(0, "0"); strings.Contains(out, "\x1b") {
		t.Fatalf("priority 0 should be plain, got %q", out)
	}
}

func restoreColorDefaults() {
	colorMode = ColorAuto
	colorNoColor = false
	outputTarget = os.Stdout
	color.NoColor = false
}
