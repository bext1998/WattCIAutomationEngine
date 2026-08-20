package main

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
)

var errFakeNotAConsole = errors.New("fake: not a console")

func TestSelectBrandPlan_NonFileWriter_UsesUnicodeUTF8(t *testing.T) {
	var buf bytes.Buffer
	got := selectBrandPlan(&buf)
	if got != planUnicodeUTF8 {
		t.Fatalf("got %v, want planUnicodeUTF8", got)
	}
}

func TestSelectBrandPlanWithProbe_ConsoleModeFails_UsesUnicodeUTF8(t *testing.T) {
	probe := consoleProbe{
		getConsoleMode: func(*os.File) error { return errFakeNotAConsole },
		getCurrentConsoleFontEx: func(*os.File) (consoleFontInfoEx, bool) {
			t.Fatal("should not be called when GetConsoleMode fails")
			return consoleFontInfoEx{}, false
		},
	}
	got := selectBrandPlanWithProbe(os.Stdout, probe)
	if got != planUnicodeUTF8 {
		t.Fatalf("got %v, want planUnicodeUTF8", got)
	}
}

func TestSelectBrandPlanWithProbe_FontQueryFails_UsesASCII(t *testing.T) {
	probe := consoleProbe{
		getConsoleMode: func(*os.File) error { return nil },
		getCurrentConsoleFontEx: func(*os.File) (consoleFontInfoEx, bool) {
			return consoleFontInfoEx{}, false
		},
	}
	got := selectBrandPlanWithProbe(os.Stdout, probe)
	if got != planASCIIConsole {
		t.Fatalf("got %v, want planASCIIConsole", got)
	}
}

func TestSelectBrandPlanWithProbe_RasterFont_UsesASCII(t *testing.T) {
	probe := consoleProbe{
		getConsoleMode: func(*os.File) error { return nil },
		getCurrentConsoleFontEx: func(*os.File) (consoleFontInfoEx, bool) {
			return consoleFontInfoEx{FontFamily: 0}, true // TMPF_TRUETYPE bit 不存在
		},
	}
	got := selectBrandPlanWithProbe(os.Stdout, probe)
	if got != planASCIIConsole {
		t.Fatalf("got %v, want planASCIIConsole", got)
	}
}

func TestSelectBrandPlanWithProbe_TrueTypeFont_UsesUnicodeConsole(t *testing.T) {
	probe := consoleProbe{
		getConsoleMode: func(*os.File) error { return nil },
		getCurrentConsoleFontEx: func(*os.File) (consoleFontInfoEx, bool) {
			return consoleFontInfoEx{FontFamily: tmpfTrueType}, true
		},
	}
	got := selectBrandPlanWithProbe(os.Stdout, probe)
	if got != planUnicodeConsole {
		t.Fatalf("got %v, want planUnicodeConsole", got)
	}
}

func TestBrandBlock_ASCIIVersionIsPureASCII(t *testing.T) {
	block := brandBlock(asciiBanner, asciiTagline, asciiHint)
	for i, r := range block {
		if r > 127 {
			t.Fatalf("ASCII brand block contains non-ASCII rune %q at byte offset %d", r, i)
		}
	}
}

func TestBrandBlock_ContainsExpectedSections(t *testing.T) {
	block := brandBlock(unicodeBanner, unicodeTagline, unicodeHint)
	if !strings.Contains(block, unicodeBanner) {
		t.Fatal("missing banner")
	}
	if !strings.Contains(block, unicodeTagline) {
		t.Fatal("missing tagline")
	}
	if !strings.Contains(block, unicodeHint) {
		t.Fatal("missing hint")
	}
}
