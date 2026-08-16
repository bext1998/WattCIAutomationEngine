# Watt — Repo 結構地圖

> 產出日期：2026-08-10
> 工具：Codex（maze-repo-map）
> 最後更新：2026-08-16（Issue #2～#9、#24、#30 / PR #21～#23、#25～#28、#33～#35 合併後；Phase 1 Must Have 全數完成，對抗式審查回溯進行中）

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
  docs/agent-rules/      — 按任務按需讀取的規則檔（process、result-env、testing、github-actions）
  cmd/watt/              — CLI 入口（已實作骨架）
    main.go              — main／execute()：錯誤輸出與 exit code 對應
    root.go              — cobra root command、run／check（含 --env 環境探測）子命令
    exit.go              — exit code 常數與 exitError 型別
    root_test.go         — CLI 行為測試
  internal/              — 六個核心 package，全部已有實作
  scripts/build.ps1      — Windows/amd64、CGO_ENABLED=0 build script
  .git/                  — Git metadata
```

`internal/` 各 package 的目前職責與實作狀態：

```
internal/orchestrator/   — pipeline 選取、循序執行、fail-fast、exit code（已實作，含 cancellation 三個 ctx 檢查點）
internal/pipeline/       — YAML 載入、資料模型、靜態驗證（已實作）
internal/runner/         — 單一步驟執行（exec／shell 兩種模式）、輸出擷取與狀態判定（已實作）
internal/result/         — result 組裝、序列化與寫入（已實作）
internal/env/            — host → pipeline → step env 合併與 cwd 解析、exec／shell PATH 探測、environment diagnostics 與已知值 redaction（已實作：merge.go／cwd.go／probe.go／diagnostics.go）
internal/proc/           — Windows Job Object 綁定、process tree 管理與 cancellation（已實作：job.go／proc.go／process.go）
```

---

## 關鍵檔案

| 檔案路徑 | 用途 |
|---|---|
| `docs/spec.md` | v1.3、Review 狀態的功能與介面權威規格；§7 的 Pipeline、Result、Exit Code、Process 契約為 `[FROZEN]` |
| `NEXT_ACTION.md` | 當前工作前線（Phase 1 Must Have 全數完成，下一步待使用者確認方向） |
| `cmd/watt/exit.go` | Exit code 常數（含 `EXIT_STEP_FAILED`）與 `exitError`；`EXIT_USAGE` 為 `EXIT_INVALID_PIPELINE`（2）的同值別名 |
| `cmd/watt/root.go` | Cobra root command、`run`（已接通 Exec Step）與 `check` 靜態驗證子命令 |
| `internal/pipeline/pipeline.go` | strict YAML 載入、預設 shell 與靜態驗證 |
| `internal/env/merge.go`、`internal/env/cwd.go` | 三層 env 合併與 step cwd 解析 |
| `internal/env/probe.go` | `ResolveExecutable`：包 `exec.LookPath`，供 `watt check --env` 探測 exec／shell 是否可解析 |
| `internal/orchestrator/orchestrator.go` | pipeline 選取、循序 fail-fast、result 組裝與 exit code 決定；含 Ctrl+C 對應的三個 context 檢查點（Issue #6、#8） |
| `internal/runner/runner.go` | exec／shell（pwsh、cmd）兩種模式啟動、即時輸出透傳、output_tail、env/cwd 解析（Issue #6、#7） |
| `internal/result/result.go` | Result／Step schema 落地、序列化與 `.watt/result.json` 寫入（Issue #6） |
| `internal/proc/job.go`、`internal/proc/proc.go`、`internal/proc/process.go` | Windows Job Object 建立與綁定（`CREATE_SUSPENDED` 後才 `ResumeThread`）、`TerminateJobObject` 終止整棵 process tree、cleanup 確認與 5 秒期限（A-10）兜底（Issue #8） |
| `internal/env/diagnostics.go` | `ProbeDiagnostics()`：固定探測 pwsh／cmd／bash 可用性（pwsh 版本探測有 `WaitDelay` 逾時保護）、解析選中 pipeline 的 exec 工具、列出 host 環境變數名稱；`RedactKnownValues()`：單一 pass 標記已知值匹配範圍聯集後收斂遮罩，處理重疊值（Issue #9） |
| `scripts/build.ps1` | Windows/amd64、`CGO_ENABLED=0`、`-trimpath` build；以 `-X main.version` 注入版本 |
| `DECISIONS.md` | 規格狀態與重大設計決策的索引 |
| `AGENTS.md` | 核心不可違反規則與 Routing Table；詳細規範按任務類型路由至 `docs/spec.md` 或 `docs/agent-rules/*.md` |
| `docs/agent-rules/process.md` | Process／Job Object／cancellation／shell step 的 agent 執行規則（process 相關任務時讀） |
| `docs/agent-rules/result-env.md` | Result／env 合併／redaction 的 agent 執行規則（result／env 相關任務時讀） |
| `docs/agent-rules/testing.md` | 測試與驗證要求（撰寫測試或判斷驗證是否足夠時讀） |
| `docs/agent-rules/github-actions.md` | GitHub Actions Node runtime 約束（修改 workflow 時讀） |
| `MAZE_PROJECT.md` | 專案路徑、GitHub repository 與工作流設定 |
| `PROJECT_BRIEF.md` | 專案定位、目標、技術棧與限制摘要 |
| `CLAUDE.md` | Claude Code 輕量入口，路由至 `AGENTS.md` |
| `README.md` | 最小專案簡介 |

---

## 進入點

- **啟動方式**：目前可使用 `watt --version`、`watt --help`、`watt check`、`watt check --env`、`watt run [pipeline]`。
- **主要進入點**：`cmd/watt/main.go` 的 `main()`；實際邏輯在 `execute()`，回傳 exit code 交給 `os.Exit`。
- **目前行為**：`watt --version` 輸出版本；`watt` 印 help；`check` 載入並 strict decode／靜態驗證 repo root 的 `watt.yaml`，不啟動 step；`check --env` 額外遍歷全部 pipeline／step，探測 `exec` 目標與 `run` 所需 shell（pwsh／cmd）是否可在 PATH 解析，缺失時回 `EXIT_ENVIRONMENT_UNAVAILABLE`（3）並列出缺項，不啟動任何 process、不寫 result.json；`run [pipeline]` 依 default／具名 pipeline 選取後循序 fail-fast 執行 `exec` 與 `run`（shell：pwsh／cmd）兩種型別 step，每個 step 皆綁定 Windows Job Object 後才啟動使用者程式碼，即時透傳 stdout/stderr 並寫出 `.watt/result.json`；Ctrl+C 會終止整棵 process tree，5 秒內確認清空回 `EXIT_CANCELLED`（4），確認不了回 `EXIT_INTERNAL_ERROR`（5，不謊報 cancelled）；`result.json` 的 `environment` 診斷區塊（`os`／`arch`／`shell_available`／`resolved_tools`／`env_var_names`，僅名稱不含值）與 `resolved_command`／`output_tail` 的已知環境值遮罩已落地（Issue #9）；usage error（未知命令／旗標／多餘參數）回 2。

---

## 建置

```powershell
.\scripts\build.ps1 [-OutputPath <path>] [-Version <version>]
```

固定 `GOOS=windows`、`GOARCH=amd64`、`CGO_ENABLED=0`，預設輸出 `dist\watt.exe`；script 結束時還原這三個環境變數。

---

## 測試

- **測試檔案**：`cmd/watt/root_test.go`（CLI、`run`／`check`／`check --env` 端到端、無副作用／失敗路徑、usage error、help 與 `exitError`；`TestExecStep_NoShellIndirection` 已改用即時編譯的 argv helper 真正驗證 P-4；`TestRun_FailFastStopsAfterCwdFailure`、`TestRun_ResultWriteFailureReturnsInternalError` 驗證 exit code 邊界）；`internal/pipeline/pipeline_test.go`（載入與靜態驗證）；`internal/env/*_test.go`（env 合併、cwd 解析、`ResolveExecutable` PATH 探測、`diagnostics_test.go` 涵蓋 `ProbeDiagnostics` 安全性、pwsh timeout 與 grandchild pipe-handle 情境、`RedactKnownValues` 一般與重疊值案例）；`internal/runner/runner_test.go`（exec／shell 啟動、output_tail、cwd／command 失敗、`TestResult_OutputTail_RedactsKnownEnvValues`、`TestRunKeepsOutputTailValidAtMultibyteBoundary` 驗證 R-9 多位元組截斷邊界不再無限迴圈）；`internal/result/result_test.go`（Result schema 序列化）；`internal/proc/proc_test.go`（Job Object 綁定、`TestCancel_KillsDescendantProcesses`、`TestProc_NoOrphansOnNormalCompletion`、`TestExitCode_InternalErrorOnUnconfirmedCancellation`）；`internal/orchestrator/orchestrator_test.go`（含 `TestRunCancellationReturnsCancelled`、`TestEnvDiagnostics_NoValuesLeaked` 端到端）。
- **執行測試**：`go test ./...`；repository 現有 `.github/workflows/ci.yml`（push／PR／手動觸發，Windows runner 執行 go vet／go test／build／smoke test），已在 PR #25～#28、#33～#35 與 main push 各成功執行一次，但仍沒有獨立 QA 報告；取消相關時序敏感測試另跑過 `-count=3` 確認穩定（本機驗證，未進 CI）；PR #34 的 pwsh timeout 修正、PR #35 的 R-9 無限迴圈修正皆另外做過反向驗證（本機暫時還原修正前邏輯重跑測試，確認會如預期卡死／失敗，證明測試非假綠燈；均已還原）。

---

## 設定

- **環境設定**：`go.mod`／`go.sum` 已建立；尚無 `.env` 或 `watt.yaml`。
- **關鍵設定項**：規格定義 pipeline 預設檔案為 repo root 的 `watt.yaml`，結果預設寫入 `.watt/result.json`；pipeline 載入／驗證、exec／shell step 執行、result 寫入與 `environment` 診斷／redaction（Issue #6、#7、#8、#9）皆已完成——Phase 1 Must Have（§5.1）全數落地。
- **新增依賴**：`golang.org/x/sys/windows`（Issue #8 引入；標準庫不提供 Job Object API，`os.Process` 也不對外公開子行程原生 handle，無法在不繞過 `exec.Cmd` 的情況下滿足 P-1 時序要求）。

---

## 目前未知項目

- Issue #2～#9、#24（Phase 1 Must Have）與 #30（Exec Step 對抗式審查）皆已完成並關閉；[Issue #29](https://github.com/bext1998/WattCIAutomationEngine/issues/29)（對抗式審查回溯總覽）**尚未關閉**，[#31](https://github.com/bext1998/WattCIAutomationEngine/issues/31)（Env 三層合併／cwd 解析，P1，security）是目前工作前線，其後接 [#32](https://github.com/bext1998/WattCIAutomationEngine/issues/32)（Pipeline 資料模型，P2）。GitHub repo 現已建立 P0～P4 優先級標籤（`MAZE_PROJECT.md` 先前記錄的「無標籤」已過期）。
- **PR #35（Issue #30）修正了一個真實的嚴重 bug**：`internal/runner` 的 `tailBuffer.String()`（R-9 output_tail 截斷）在截斷點恰好落在多位元組 UTF-8 字元中間時會無限迴圈，卡死 `watt run`；已修正並經獨立 subagent 複審（含數學推導與實際複現），結論 go。
- `watt.yaml` 尚未納入 repository；目前只能透過測試或外部工作目錄提供 pipeline 定義驗證 `watt check`／`watt run`。
- repository 現有 `.github/workflows/ci.yml`（push／PR／手動觸發，Windows runner 執行 go vet／go test／build／smoke test），已在 PR #25～#28、#33～#35 與 main push 各成功執行一次；仍沒有獨立 QA 報告。
- Issue #9／PR #34 已知殘餘風險：pwsh 版本探測的 `WaitDelay` 修正保證探測呼叫本身會逾時返回，但不會終止造成卡住的 wrapper 之 grandchild 行程本身（該情境下 grandchild 可能短暫繼續在背景執行）；不屬於 spec §7.4 Job Object 契約範圍。
- Issue #8／PR #33 已知驗證缺口：AC-6 checkbox 關閉當下未勾選——取消流程以 `context` cancellation 模擬 Ctrl+C（語意等價），未實際發送 OS 層級 Ctrl+C 訊號端到端驗證，CI 也未針對取消情境跑專門驗證（僅本機 Windows 手動測過）；詳見 `NEXT_ACTION.md`。
- PR #27（Issue #6）技術債（詳見 [PR #27 留言](https://github.com/bext1998/WattCIAutomationEngine/pull/27#issuecomment-5285418121)）三項測試缺口已由 PR #35 補齊兩項（`result.Write` 失敗路徑、多位元組 UTF-8 截斷邊界）；剩餘 `orchestrator.WattVersion` 寫死 `"dev"`、`resolved_command` 空白 join 對含空白參數失真兩項，經 PR #35 審查確認純屬 diagnostic 顯示問題、不影響實際 argv 或語意判定，維持低優先級；PR #28（Issue #7）留下的非阻擋觀察（`TestMissingPwsh_NoFallbackTo51` 覆蓋不足、`runner.shellArgs()` 防禦深度缺口）尚未處理。
- 對抗式審查回溯進度：[Issue #29](https://github.com/bext1998/WattCIAutomationEngine/issues/29) 總覽尚未關閉，#30 已完成，#31（P1）為下一前線，#32（P2）待其後。
- PR #35 審查過程中額外發現：`docs/spec.md` §4.2 提到的 `--output json` 行為目前 repo 尚未實作，對應既有 [Issue #10](https://github.com/bext1998/WattCIAutomationEngine/issues/10)（P3，非阻塞，純記錄）。
