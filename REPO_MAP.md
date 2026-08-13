# Watt — Repo 結構地圖

> 產出日期：2026-08-10
> 工具：Codex（maze-repo-map）
> 最後更新：2026-08-12（Issue #2～#5、#24 / PR #21～#23、#25、#26 合併後）

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
  .github/workflows/ci.yml — CI：push／PR／手動觸發，windows-latest 執行 go vet／go test／build／smoke test
  docs/spec.md           — Watt Phase 1 v1.3 功能、架構與驗收規格
  cmd/watt/              — CLI 入口（已實作骨架）
    main.go              — main／execute()：錯誤輸出與 exit code 對應
    root.go              — cobra root command、run／check（含 --env 環境探測）子命令
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
internal/env/            — host → pipeline → step env 合併與 cwd 解析、exec／shell PATH 探測（已實作）
internal/proc/           — 規劃中的 Windows Job Object 與 process tree 管理（骨架）
```

---

## 關鍵檔案

| 檔案路徑 | 用途 |
|---|---|
| `docs/spec.md` | v1.3、Review 狀態的功能與介面權威規格；§7 的 Pipeline、Result、Exit Code、Process 契約為 `[FROZEN]` |
| `NEXT_ACTION.md` | 當前工作前線（Issue #9：Environment Diagnostics 與已知環境值 Redaction） |
| `cmd/watt/exit.go` | Exit code 常數（含 `EXIT_STEP_FAILED`）與 `exitError`；`EXIT_USAGE` 為 `EXIT_INVALID_PIPELINE`（2）的同值別名 |
| `cmd/watt/root.go` | Cobra root command、`run`（已接通 Exec Step）與 `check` 靜態驗證子命令 |
| `internal/pipeline/pipeline.go` | strict YAML 載入、預設 shell 與靜態驗證 |
| `internal/env/merge.go`、`internal/env/cwd.go` | 三層 env 合併與 step cwd 解析 |
| `internal/env/probe.go` | `ResolveExecutable`：包 `exec.LookPath`，供 `watt check --env` 探測 exec／shell 是否可解析 |
| `internal/orchestrator/orchestrator.go` | pipeline 選取、循序 fail-fast、result 組裝與 exit code 決定（Issue #6） |
| `internal/runner/runner.go` | Exec 模式直接啟動、即時輸出透傳、output_tail、env/cwd 解析（Issue #6） |
| `internal/result/result.go` | Result／Step schema 落地、序列化與 `.watt/result.json` 寫入（Issue #6） |
| `scripts/build.ps1` | Windows/amd64、`CGO_ENABLED=0`、`-trimpath` build；以 `-X main.version` 注入版本 |
| `DECISIONS.md` | 規格狀態與重大設計決策的索引 |
| `AGENTS.md` | 專案範圍、架構方向、執行／authoring 權限與驗證規則 |
| `MAZE_PROJECT.md` | 專案路徑、GitHub repository 與工作流設定 |
| `PROJECT_BRIEF.md` | 專案定位、目標、技術棧與限制摘要 |
| `CLAUDE.md` | Claude Code 工作規則與遠端同步規則 |
| `README.md` | 最小專案簡介 |

---

## 進入點

- **啟動方式**：目前可使用 `watt --version`、`watt --help`、`watt check`、`watt check --env`、`watt run [pipeline]`。
- **主要進入點**：`cmd/watt/main.go` 的 `main()`；實際邏輯在 `execute()`，回傳 exit code 交給 `os.Exit`。
- **目前行為**：`watt --version` 輸出版本；`watt` 印 help；`check` 載入並 strict decode／靜態驗證 repo root 的 `watt.yaml`，不啟動 step；`check --env` 額外遍歷全部 pipeline／step，探測 `exec` 目標與 `run` 所需 shell（pwsh／cmd）是否可在 PATH 解析，缺失時回 `EXIT_ENVIRONMENT_UNAVAILABLE`（3）並列出缺項，不啟動任何 process、不寫 result.json；`run [pipeline]` 依 default／具名 pipeline 選取後循序 fail-fast 執行 `exec` 型 step，即時透傳 stdout/stderr 並寫出 `.watt/result.json`，`run` 型（shell）step 仍回報未實作失敗；usage error（未知命令／旗標／多餘參數）回 2。

---

## 建置

```powershell
.\scripts\build.ps1 [-OutputPath <path>] [-Version <version>]
```

固定 `GOOS=windows`、`GOARCH=amd64`、`CGO_ENABLED=0`，預設輸出 `dist\watt.exe`；script 結束時還原這三個環境變數。

---

## 測試

- **測試檔案**：`cmd/watt/root_test.go`（CLI、`run`／`check`／`check --env` 端到端、無副作用／失敗路徑、usage error、help 與 `exitError`）；`internal/pipeline/pipeline_test.go`（載入與靜態驗證）；`internal/env/*_test.go`（env 合併、cwd 解析與 `ResolveExecutable` PATH 探測）；`internal/runner/runner_test.go`（Exec 啟動、output_tail、cwd／command 失敗）；`internal/result/result_test.go`（Result schema 序列化）。
- **執行測試**：`go test ./...`；本次 closeout 未新增執行 QA，repository 現有 `.github/workflows/ci.yml`（push／PR／手動觸發，Windows runner 執行 go vet／go test／build／smoke test），已在 PR #25、#26、#27 與 main push 各成功執行一次，但仍沒有獨立 QA 報告。

---

## 設定

- **環境設定**：`go.mod`／`go.sum` 已建立；尚無 `.env` 或 `watt.yaml`。
- **關鍵設定項**：規格定義 pipeline 預設檔案為 repo root 的 `watt.yaml`，結果預設寫入 `.watt/result.json`；pipeline 載入／驗證與 result 寫入（Issue #6）皆已完成，`environment` 診斷區塊與已知環境值 redaction 待 Issue #9。

---

## 目前未知項目

- Issue #6（Exec Step／`watt run`）已完成並關閉（PR #27）；Issue #9（Environment Diagnostics／已知環境值 Redaction）已解除前置阻塞，是目前工作前線；#7（Shell Step）、#8（Job Object／cancellation）為 P1，亦已解除前置阻塞。
- `watt.yaml` 尚未納入 repository；目前只能透過測試或外部工作目錄提供 pipeline 定義驗證 `watt check`／`watt run`。
- repository 現有 `.github/workflows/ci.yml`（push／PR／手動觸發，Windows runner 執行 go vet／go test／build／smoke test），已在 PR #25、#26、#27 與 main push 各成功執行一次；仍沒有獨立 QA 報告。目前 GitHub 的 #21～#23、#25～#27 PR 已合併，可引用其 CI 執行紀錄作為 rollup。
- PR #27（Issue #6）審查留下的非阻擋技術債（詳見 [PR #27 留言](https://github.com/bext1998/WattCIAutomationEngine/pull/27#issuecomment-5285418121) 與 `NEXT_ACTION.md`）：`orchestrator.WattVersion` 寫死 `"dev"`、`resolved_command` 空白 join 對含空白參數失真、Run-mode 失敗訊息未帶 step 名稱、`result.Write` 失敗時吞掉原始 step 錯誤，以及三項測試缺口。
