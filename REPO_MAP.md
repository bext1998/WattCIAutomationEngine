# Watt — Repo 結構地圖

> 產出日期：2026-08-10
> 工具：Codex（maze-repo-map）
> 最後更新：2026-08-10（Issue #2～#4 / PR #21～#23 合併後）

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
  internal/              — 六個核心 package（pipeline、env 已有基礎實作，其餘仍為骨架）
  scripts/build.ps1      — Windows/amd64、CGO_ENABLED=0 build script
  .git/                  — Git metadata
```

`internal/` 各 package 的目前職責與實作狀態：

```
internal/orchestrator/   — 規劃中的 pipeline 選取、循序執行、fail-fast、exit code（骨架）
internal/pipeline/       — YAML 載入、資料模型、靜態驗證（已實作）
internal/runner/         — 規劃中的單一步驟執行、輸出擷取與狀態判定（骨架）
internal/result/         — 規劃中的 result 組裝、序列化與寫入（骨架）
internal/env/            — host → pipeline → step env 合併與 cwd 解析（已實作）
internal/proc/           — 規劃中的 Windows Job Object 與 process tree 管理（骨架）
```

---

## 關鍵檔案

| 檔案路徑 | 用途 |
|---|---|
| `docs/spec.md` | v1.3、Review 狀態的功能與介面權威規格；§7 的 Pipeline、Result、Exit Code、Process 契約為 `[FROZEN]` |
| `NEXT_ACTION.md` | 當前工作前線（Issue #6：Exec Step 執行核心） |
| `cmd/watt/exit.go` | Exit code 常數與 `exitError`；`EXIT_USAGE` 為 `EXIT_INVALID_PIPELINE`（2）的同值別名 |
| `cmd/watt/root.go` | Cobra root command、`run` stub 與 `check` 靜態驗證子命令 |
| `internal/pipeline/pipeline.go` | strict YAML 載入、預設 shell 與靜態驗證 |
| `internal/env/merge.go`、`internal/env/cwd.go` | 三層 env 合併與 step cwd 解析 |
| `scripts/build.ps1` | Windows/amd64、`CGO_ENABLED=0`、`-trimpath` build；以 `-X main.version` 注入版本 |
| `DECISIONS.md` | 規格狀態與重大設計決策的索引 |
| `AGENTS.md` | 專案範圍、架構方向、執行／authoring 權限與驗證規則 |
| `MAZE_PROJECT.md` | 專案路徑、GitHub repository 與工作流設定 |
| `PROJECT_BRIEF.md` | 專案定位、目標、技術棧與限制摘要 |
| `CLAUDE.md` | Claude Code 工作規則與遠端同步規則 |
| `README.md` | 最小專案簡介 |

---

## 進入點

- **啟動方式**：目前可使用 `watt --version`、`watt --help`、`watt check`；`watt run [pipeline]` 與 `watt check --env` 為後續實作。
- **主要進入點**：`cmd/watt/main.go` 的 `main()`；實際邏輯在 `execute()`，回傳 exit code 交給 `os.Exit`。
- **目前行為**：`watt --version` 輸出版本；`watt` 印 help；`check` 載入並 strict decode／靜態驗證 repo root 的 `watt.yaml`，不啟動 step；`run` 回 `EXIT_INTERNAL_ERROR`（5）並輸出 `run is not implemented`；usage error（未知命令／旗標／多餘參數）回 2。

---

## 建置

```powershell
.\scripts\build.ps1 [-OutputPath <path>] [-Version <version>]
```

固定 `GOOS=windows`、`GOARCH=amd64`、`CGO_ENABLED=0`，預設輸出 `dist\watt.exe`；script 結束時還原這三個環境變數。

---

## 測試

- **測試檔案**：`cmd/watt/root_test.go`（CLI、`check` 無副作用／失敗路徑、usage error、help 與 `exitError`）；`internal/pipeline/pipeline_test.go`（載入與靜態驗證）；`internal/env/*_test.go`（env 合併與 cwd 解析）。
- **執行測試**：`go test ./...`；本次 closeout 未新增執行 QA，repository 現有 `.github/workflows/ci.yml`（push／PR／手動觸發，Windows runner 執行 go vet／go test／build／smoke test），但沒有獨立 QA 報告或 GitHub Actions 執行紀錄。

---

## 設定

- **環境設定**：`go.mod`／`go.sum` 已建立；尚無 `.env` 或 `watt.yaml`。
- **關鍵設定項**：規格定義 pipeline 預設檔案為 repo root 的 `watt.yaml`，結果預設寫入 `.watt/result.json`；pipeline 載入／驗證已完成，result 寫入待 Issue #6。

---

## 目前未知項目

- Issue #6（Exec Step／`watt run`）尚未開始，後續 #7、#8、#9 等執行期能力仍受其相依關係影響。
- `watt.yaml` 尚未納入 repository；目前只能透過測試或外部工作目錄提供 pipeline 定義驗證 `watt check`。
- repository 現有 `.github/workflows/ci.yml`（push／PR／手動觸發，Windows runner 執行 go vet／go test／build／smoke test），但沒有 GitHub Actions 執行紀錄或獨立 QA 報告；目前 GitHub 的 #21～#23 PR 已合併，但沒有可引用的 CI／review rollup。
