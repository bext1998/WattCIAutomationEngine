package main

import (
	"fmt"
	"io"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ---- 品牌文字常數 ----
// 這些字串內容是最終定案，不要更動任何字元，包含空白與換行。

const unicodeBanner = `██╗    ██╗ █████╗ ████████╗████████╗
██║    ██║██╔══██╗╚══██╔══╝╚══██╔══╝
██║ █╗ ██║███████║   ██║      ██║
██║███╗██║██╔══██║   ██║      ██║
╚███╔███╔╝██║  ██║   ██║      ██║
 ╚══╝╚══╝ ╚═╝  ╚═╝   ╚═╝      ╚═╝`

const unicodeTagline = "確定性的本機 Pipeline 執行引擎 · Deterministic local-first pipeline execution engine"

const unicodeHint = "已有 watt.yaml？先執行 watt check --env 確認設定與執行環境。"

const asciiBanner = ` __        ___  _____ _____
 \ \      / / \|_   _|_   _|
  \ \ /\ / / _ \ | |   | |
   \ V  V / ___ \| |   | |
    \_/\_/_/   \_\_|   |_|`

const asciiTagline = "Deterministic local-first pipeline execution engine"

const asciiHint = `Have a watt.yaml? Run "watt check --env" to verify your setup.`

// brandBlock 組出「橫幅 + 空行 + 標語 + 空行 + 提示 + 空行」的完整區塊。
// banner 參數本身不能有結尾換行（最後一行就是圖形最後一列），這個函式會補上。
func brandBlock(banner, tagline, hint string) string {
	return banner + "\n\n" + tagline + "\n\n" + hint + "\n\n"
}

// ---- Unicode／ASCII 判斷邏輯 ----

type brandPlan int

const (
	// planUnicodeUTF8：輸出目標不是真正的 Windows console（redirect、pipe、
	// 測試用 buffer 等），直接寫 Unicode 內容的 UTF-8 bytes，不做任何降級。
	// 這是刻意的政策：Watt 不應該根據自己 attached console 的字型，去猜測
	// redirect/pipe 下游的顯示能力。
	planUnicodeUTF8 brandPlan = iota
	// planUnicodeConsole：真正的 Windows console，且目前 console 字型被
	// Win32 回報為 TrueType，用 WriteConsole 寫 UTF-16。
	planUnicodeConsole
	// planASCIIConsole：真正的 Windows console，但字型查詢失敗或字型是
	// raster／點陣字型（例如舊式 "Terminal" 字型），保守降級。
	planASCIIConsole
)

// CONSOLE_FONT_INFOEX（Win32 struct，欄位順序與型別必須完全對應，不要更動）。
type consoleFontInfoEx struct {
	CbSize      uint32
	NFont       uint32
	DwFontSizeX int16
	DwFontSizeY int16
	FontFamily  uint32
	FontWeight  uint32
	FaceName    [32]uint16
}

// TMPF_TRUETYPE：FontFamily 欄位的 bit flag，代表目前 console 字型是
// TrueType（而不是 raster／點陣字型）。這不是「所有 glyph 都存在」的證明，
// 只是一個保守、可重現的 heuristic——Windows 沒有更可靠的跨宿主 API。
const tmpfTrueType = 0x04

var (
	kernel32DLL                 = windows.NewLazySystemDLL("kernel32.dll")
	procGetCurrentConsoleFontEx = kernel32DLL.NewProc("GetCurrentConsoleFontEx")
)

// getCurrentConsoleFontExRaw 直接呼叫 Win32 GetCurrentConsoleFontEx。
// golang.org/x/sys/windows 沒有直接包裝這個 API，用 LazyDLL 呼叫，不需要
// 新增任何外部依賴。
func getCurrentConsoleFontExRaw(handle windows.Handle) (consoleFontInfoEx, bool) {
	var info consoleFontInfoEx
	info.CbSize = uint32(unsafe.Sizeof(info))
	r, _, _ := procGetCurrentConsoleFontEx.Call(
		uintptr(handle),
		0, // bMaximumWindow = FALSE
		uintptr(unsafe.Pointer(&info)),
	)
	return info, r != 0
}

// consoleProbe 把「查 console mode」「查 console 字型」包成可注入的函式，
// 讓測試可以用假資料覆蓋，不必依賴開發機實際字型，也不能呼叫
// SetConsoleOutputCP 改變共享 console 狀態。
type consoleProbe struct {
	getConsoleMode          func(f *os.File) error // 成功回 nil，失敗回 err
	getCurrentConsoleFontEx func(f *os.File) (consoleFontInfoEx, bool)
}

var defaultConsoleProbe = consoleProbe{
	getConsoleMode: func(f *os.File) error {
		var mode uint32
		return windows.GetConsoleMode(windows.Handle(f.Fd()), &mode)
	},
	getCurrentConsoleFontEx: func(f *os.File) (consoleFontInfoEx, bool) {
		return getCurrentConsoleFontExRaw(windows.Handle(f.Fd()))
	},
}

// selectBrandPlanWithProbe 是實際判斷邏輯，probe 參數讓測試可以注入假資料。
func selectBrandPlanWithProbe(w io.Writer, probe consoleProbe) brandPlan {
	f, ok := w.(*os.File)
	if !ok {
		return planUnicodeUTF8
	}
	if err := probe.getConsoleMode(f); err != nil {
		// redirect、pipe、或其他非 console 的 *os.File。
		return planUnicodeUTF8
	}
	info, ok := probe.getCurrentConsoleFontEx(f)
	if !ok {
		return planASCIIConsole
	}
	if info.FontFamily&tmpfTrueType == 0 {
		return planASCIIConsole
	}
	return planUnicodeConsole
}

func selectBrandPlan(w io.Writer) brandPlan {
	return selectBrandPlanWithProbe(w, defaultConsoleProbe)
}

// ---- 輸出 ----

// writeUTF16ToConsole 把 text 轉成 UTF-16 後用 windows.WriteConsole 寫入，
// 繞過 console output code page 的限制。
func writeUTF16ToConsole(f *os.File, text string) error {
	utf16Text, err := windows.UTF16FromString(text)
	if err != nil {
		return err
	}
	// UTF16FromString 會補一個 NUL 結尾，寫入前要去掉，否則畫面上會多印出
	// 一個 NUL 字元。
	if n := len(utf16Text); n > 0 && utf16Text[n-1] == 0 {
		utf16Text = utf16Text[:n-1]
	}
	if len(utf16Text) == 0 {
		return nil
	}
	handle := windows.Handle(f.Fd())
	var written uint32
	return windows.WriteConsole(handle, &utf16Text[0], uint32(len(utf16Text)), &written, nil)
}

// writeBrand 依 selectBrandPlan 的判斷結果，把品牌橫幅／標語／提示寫到 out。
func writeBrand(out io.Writer) {
	switch selectBrandPlan(out) {
	case planUnicodeConsole:
		text := brandBlock(unicodeBanner, unicodeTagline, unicodeHint)
		if f, ok := out.(*os.File); ok {
			if err := writeUTF16ToConsole(f, text); err == nil {
				return
			}
			// WriteConsole 寫入失敗時，退回一般 UTF-8 bytes 寫入，至少不要
			// 讓裸執行 watt 因為橫幅寫入失敗而整個中斷。
		}
		fmt.Fprint(out, text)
	case planASCIIConsole:
		fmt.Fprint(out, brandBlock(asciiBanner, asciiTagline, asciiHint))
	default: // planUnicodeUTF8
		fmt.Fprint(out, brandBlock(unicodeBanner, unicodeTagline, unicodeHint))
	}
}
