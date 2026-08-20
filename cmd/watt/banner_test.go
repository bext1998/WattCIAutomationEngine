package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
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

func TestWriteBrandWithPlan_ASCIIConsole_WritesExactBlock(t *testing.T) {
	var buf bytes.Buffer
	err := writeBrandWithPlan(&buf, planASCIIConsole, func(*os.File, string) error {
		t.Fatal("consoleWrite should not be called for planASCIIConsole")
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := brandBlock(asciiBanner, asciiTagline, asciiHint)
	if got := buf.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWriteBrandWithPlan_UnicodeUTF8_WritesExactBlock(t *testing.T) {
	var buf bytes.Buffer
	err := writeBrandWithPlan(&buf, planUnicodeUTF8, func(*os.File, string) error {
		t.Fatal("consoleWrite should not be called for planUnicodeUTF8")
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := brandBlock(unicodeBanner, unicodeTagline, unicodeHint)
	if got := buf.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWriteBrandWithPlan_UnicodeConsole_PassesExactTextToConsoleWrite(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "watt-brand-test")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer f.Close()

	var gotFile *os.File
	var gotText string
	err = writeBrandWithPlan(f, planUnicodeConsole, func(w *os.File, text string) error {
		gotFile = w
		gotText = text
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotFile != f {
		t.Error("consoleWrite did not receive the same *os.File")
	}
	want := brandBlock(unicodeBanner, unicodeTagline, unicodeHint)
	if gotText != want {
		t.Errorf("got %q, want %q", gotText, want)
	}
}

func TestWriteBrandWithPlan_UnicodeConsole_PropagatesConsoleWriteError(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "watt-brand-test")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer f.Close()

	sentinel := errors.New("fake console write failure")
	err = writeBrandWithPlan(f, planUnicodeConsole, func(*os.File, string) error {
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Errorf("got error %v, want %v", err, sentinel)
	}
}

func TestWriteUTF16ToConsole_ShortWrite_WritesRemainder(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "watt-brand-test")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer f.Close()

	original := consoleWriteOnce
	defer func() { consoleWriteOnce = original }()

	var calls int
	var gotChunks []string
	consoleWriteOnce = func(handle windows.Handle, buf *uint16, toWrite uint32, written *uint32) error {
		calls++
		if calls == 1 {
			// 第一次只假裝寫了一半。
			*written = toWrite / 2
		} else {
			*written = toWrite
		}
		// 把這次實際回報寫入的 UTF-16 buffer 轉回字串記錄下來，方便比對。
		chunk := unsafe.Slice(buf, *written)
		gotChunks = append(gotChunks, windows.UTF16ToString(chunk))
		return nil
	}

	text := "hello short write"
	if err := writeUTF16ToConsole(f, text); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls < 2 {
		t.Fatalf("expected at least 2 WriteConsole calls for a short write, got %d", calls)
	}
	joined := strings.Join(gotChunks, "")
	if joined != text {
		t.Errorf("reassembled written text = %q, want %q", joined, text)
	}
}

func TestWriteUTF16ToConsole_ZeroWrite_ReturnsShortWriteError(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "watt-brand-test")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer f.Close()

	original := consoleWriteOnce
	defer func() { consoleWriteOnce = original }()

	consoleWriteOnce = func(handle windows.Handle, buf *uint16, toWrite uint32, written *uint32) error {
		*written = 0
		return nil
	}

	err = writeUTF16ToConsole(f, "hello")
	if !errors.Is(err, io.ErrShortWrite) {
		t.Errorf("got error %v, want io.ErrShortWrite", err)
	}
}
