# Watt — 下一步行動

> 僅保留當前有效前線；明確 closeout 時整體重建，不追加歷史。

## 目前狀態

`blocked`（部分前線）。最近完成並合併：Issue #50（裸執行 `watt` 顯示 ASCII 品牌橫幅／雙語標語／新手上路提示，PR #51，spec.md v1.5 → v1.6）。實作由 spatula-otter（codex）依逐字規格完成，技術方案（Unicode／ASCII 判定演算法、`WriteConsole` short write 處理）經技術顧問 whisk-badger（pi）兩輪審查（REVISE → 修正 → GO）；Claude Code 每輪都獨立重跑驗證，未只採信回報。此前完成：Issue #44（Release workflow + `watt_version` 硬編碼修正，PR #45）、Issue #46（新增 Watt Agent Skill）。文件站三個 Issue（#47/#48/#49，P3，documentation）仍未動工：工作目錄現有**未 commit** 的 `site/index.html`、`site/assets/`，疑似 #48（最小網站骨架）已起步，但 #48 明確依賴 #47（內容架構與單一事實來源決策），而 #47 目前沒有任何決策記錄——**#48 仍處於 blocked**，不建議在 #47 補齊前繼續往下堆 `site/`。獨立於文件站之外，Issue #37（P2／security，env key 大小寫檢查）無阻塞、隨時可開始。Parent [Issue #1](https://github.com/bext1998/WattCIAutomationEngine/issues/1)（Phase 1 MVP 整體追蹤）仍開啟。

## 下一個 Session 目標

在「先做無阻塞的 #37」與「先補 #47 決策記錄解開 #48」之間，由使用者選擇下一個工作前線。

## 行動（最多 3 項）

1. 跟使用者確認：先做 #37，還是先處理 #47。
2. 若選 #37：讀 Issue #37、`internal/pipeline/pipeline.go`（`Validate()`）與 `internal/env`（Issue #31／PR #36 對同層大小寫重複 key 的既有處理方式），補驗證規則與回歸測試。
3. 若選 #47：讀 Issue #47 全文與 #48／#49，產出內容架構決策記錄（`docs/spec.md`／`README.md`／網站三者的權威邊界），再回頭檢視 `site/` 現有草稿是否符合決策。

## 阻塞與待決策

- 待決策：#37 vs #47（解開 #48 阻塞）由使用者選擇下一個工作前線。
- #48 blocked by #47（無決策記錄）；`site/` 未 commit 草稿用途待使用者確認去留。
- 已知測試缺口（延續自 PR #26／#27／#28，尚未補測，近期無 PR 觸及）：`watt check --env` 未覆蓋「PATH 有 `powershell.exe`（5.1）但無 `pwsh.exe`（7）」情境；`runner.shellArgs()` 對非 `pwsh`/`cmd` 值目前僅靠 `pipeline.Validate()` 擋住，`runner` 本身無防禦（見 `REPO_MAP.md`「目前未知項目」）。
- Issue #9／PR #34 殘餘風險：pwsh 版本探測的 `WaitDelay` 修正保證探測呼叫本身逾時返回，但不會終止卡住的 wrapper 之 grandchild 行程本身，影響範圍小，暫不視為阻塞。
- Issue #50／PR #51 已知未驗證項目：ASCII 降級分支（真正 console＋raster／點陣字型）只靠單元測試以注入假資料驗證判定邏輯，未在真正的舊版點陣字型 console 環境跑過真實案例（Windows 無百分之百可靠的字型能力偵測 API，設計階段已承認此限制）。

## 權威連結

- `docs\spec.md`（v1.6，Review）
- [Parent Issue #1](https://github.com/bext1998/WattCIAutomationEngine/issues/1)（追蹤全部 Sub-issue 進度，以 GitHub 為工作狀態權威）
- [Issue #37](https://github.com/bext1998/WattCIAutomationEngine/issues/37)（候選前線，P2／security，無阻塞）
- [Issue #47](https://github.com/bext1998/WattCIAutomationEngine/issues/47)（候選前線，P3，阻塞 #48）／[#48](https://github.com/bext1998/WattCIAutomationEngine/issues/48)（blocked）／[#49](https://github.com/bext1998/WattCIAutomationEngine/issues/49)（依賴 #47／#48）
- [PR #51](https://github.com/bext1998/WattCIAutomationEngine/pull/51)（Issue #50，最近合併）、PR #45（Issue #44，Release workflow）、PR #46（Issue #46，Watt Agent Skill）
