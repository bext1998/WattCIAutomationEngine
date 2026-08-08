# Watt — Repo 結構地圖

> 產出日期：2026-08-09
> 工具：Codex（maze-repo-map）
> 最後更新：2026-08-09（Issue #2 / PR #21 合併後）

---

## 技術棧

- **語言**：Go 1.24.0（`go.mod`）
- **框架**：Cobra CLI v1.10.2（`github.com/spf13/cobra`）
- **資料存儲**：檔案系統；預設結果為 `.watt/result.json`
- **部署**：Windows amd64；單一 `watt.exe`、static build、zero CGO

---

## 目錄結構

```
WattCIAutomationEngine/
  AGENTS.md              — Codex／專案工作規則
  CLAUDE.md              — Coding agent 指令
  DECISIONS.md           — 有效重大決策索引
  LICENSE                — MIT 授權
  MAZE_PROJECT.md        — 專案定位與工作流設定
  NEXT_ACTION.md         — 當前工作前線
  PROJECT_BRIEF.md       — 專案背景與技術棧摘要
  README.md              — 專案簡介
  REPO_MAP.md            — 本檔；repo 結構地圖
  go.mod / go.sum        — Go module 定義與依賴鎖定
  .gitignore             — 忽略 /dist/ 與 *.exe
  docs/spec.md           — Watt Phase 1 v1.3 功能、架構與驗收規格
  cmd/watt/              — CLI 入口（已實作骨架）
    main.go              — main／execute()：錯誤輸出與 exit code 對應
    root.go              — cobra root command、run／check 子命令
    exit.go              — exit code 常數與 exitError 型別
    root_test.go         — CLI 行為測試
  internal/              — 六個核心 package（目前皆僅有 doc.go 骨架）
  scripts/build.ps1      — Windows/amd64、CGO_ENABLED=0 build script
  .git/                  — Git metadata
```

`internal/` 各 package 的規劃職責（實作尚未開始，僅有 `doc.go`）：

```
internal/orchestrator/   — pipeline 選取、循序執行、fail-fast、exit code
internal/pipeline/       — YAML 載入、資料模型、靜態驗證
internal/runner/         — 單一步驟執行、輸出擷取與狀態判定
internal/result/         — result 組裝、序列化與寫入
internal/env/            — host → pipeline → step env 合併與 diagnostics
internal/proc/           — Windows Job Object 與 process tree 管理
```

---

## 關鍵檔案

| 檔案路徑 | 用途 |
|---|---|
| `docs/spec.md` | v1.3、Review 狀態的功能與介面權威規格；§7 的 Pipeline、Result、Exit Code、Process 契約為 `[FROZEN]` |
| `NEXT_ACTION.md` | 當前工作前線（Issue #2 已完成，待下次 closeout 重建）|
| `cmd/watt/exit.go` | Exit code 常數與 `exitError`；`EXIT_USAGE` 為 `EXIT_INVALID_PIPELINE`（2）的同值別名 |
| `scripts/build.ps1` | Windows/amd64、`CGO_ENABLED=0`、`-trimpath` build；以 `-X main.version` 注入版本 |
| `DECISIONS.md` | 規格狀態與重大設計決策的索引 |
| `AGENTS.md` | 專案範圍、架構方向、執行／authoring 權限與驗證規則 |
| `MAZE_PROJECT.md` | 專案路徑、GitHub repository 與工作流設定 |
| `PROJECT_BRIEF.md` | 專案定位、目標、技術棧與限制摘要 |
| `CLAUDE.md` | Claude Code 工作規則與遠端同步規則 |
| `README.md` | 最小專案簡介 |

---

## 進入點

- **啟動方式**：規格預期以 `watt run [pipeline]`、`watt check [--env]` 執行。
- **主要進入點**：`cmd/watt/main.go` 的 `main()`；實際邏輯在 `execute()`，回傳 exit code 交給 `os.Exit`。
- **目前行為**：`watt --version` 輸出版本；`watt` 印 help；`run`／`check` 為 stub，回 `EXIT_INTERNAL_ERROR`（5）並輸出 `<name> is not implemented`；usage error（未知命令／旗標／多餘參數）回 2。

---

## 建置

```powershell
.\scripts\build.ps1 [-OutputPath <path>] [-Version <version>]
```

固定 `GOOS=windows`、`GOARCH=amd64`、`CGO_ENABLED=0`，預設輸出 `dist\watt.exe`；script 結束時還原這三個環境變數。

---

## 測試

- **測試檔案**：`cmd/watt/root_test.go`（CLI 版本輸出、stub exit code、usage error 分類、help 內容、`exitError` 包裝語意）。
- **執行測試**：`go test ./...`；`internal/` 六個 package 目前尚無測試檔。

---

## 設定

- **環境設定**：`go.mod`／`go.sum` 已建立；尚無 `.env` 或 `watt.yaml`。
- **關鍵設定項**：規格定義 pipeline 預設檔案為 repo root 的 `watt.yaml`，結果預設寫入 `.watt/result.json`；兩者尚未建立（待 Issue #3 起實作）。

---

## 目前未知項目

- `NEXT_ACTION.md` 內容停留在 Issue #2 開始前，依專案規則僅於明確 closeout 時重建，故未同步。
- `PROJECT_BRIEF.md` 與 `CLAUDE.md` 仍保留「Draft」字樣；`docs/spec.md`、`NEXT_ACTION.md` 與 `MAZE_PROJECT.md` 已標示 v1.3 為 Review，依專案規則以 `docs/spec.md` 為功能權威。
- GitHub Issues #1–#18 的 `spec-revision` metadata 仍指向 §7.3 revision 前的 spec hash；該次為純增補、未改變任何 Issue 範圍，故刻意未同步。
