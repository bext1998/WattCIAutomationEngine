# Watt — Codex 專案指令

## 專案定位

Watt 是 Windows-first、local-first 的確定性 Pipeline 執行與驗證引擎。Phase 1 以 Go 1.24.x 實作，目標產物為單一 `watt.exe`，static build、`CGO_ENABLED=0`，僅支援 `windows/amd64`。

Watt 只負責載入、驗證與依序執行 pipeline，並產出機器可解析的結果。不得讓核心認知 Taylor、看板、coding agent 或其他上層消費者；整合方向永遠是「消費者 → Watt」。Watt 不提供也不得宣稱提供 sandbox、filesystem isolation、secret store、remote runner 或 GitHub Actions 相容層。

## 開始工作前

1. 先讀 `NEXT_ACTION.md` 確認目前工作前線，再讀任務指定的 GitHub Issue 或其他驗收條件。
2. 讀取與變更直接相關的 `docs/spec.md` 章節、現有實作、型別、測試與呼叫者；不得依摘要猜測契約。
3. `docs/spec.md` 是功能、架構與驗收標準的權威來源；`DECISIONS.md` 是有效重大決策索引；`PROJECT_BRIEF.md` 與 `MAZE_PROJECT.md` 僅提供專案背景與工作流定位。
4. 文件若互相矛盾，以較新且更具體的 `docs/spec.md` 為準；若矛盾會改變公開行為、資料格式或架構方向，停止實作並向使用者確認。
5. 工作區可能含使用者或其他代理的未提交修改。編輯前執行 `git status --short` 並檢查相關 diff；不得覆蓋、還原或順手整理範圍外變更。

## 規格契約

- `docs/spec.md` §7 全部標記為 `[FROZEN]`。Pipeline 資料模型、Result Schema、Exit Code Contract 與 Process 管理契約不得自行增刪欄位、改名、改值或改變語意；任何變更都必須走 spec revision 並取得使用者確認。
- Phase 1 僅實作 §5.1 Must Have。Should Have 只有在任務明確要求時實作；Could Have、Won't Have 與 §14 列出的 Phase 2 功能不得預作空模組、介面、設定或擴充點。
- 不得擴大 Watt 的確定性宣稱。僅 step 順序、fail-fast、env 合併優先序、result schema 與 exit code mapping 屬於保證範圍；外部 command、網路、時間、耗時及專案自身結果不屬於保證範圍。
- 公開 CLI、YAML、JSON、exit code 與可觀察行為預設保持相容。實作與規格不一致時，修正實作，不得暗改規格來配合程式碼。

## 架構與依賴方向

維持下列模組責任與單向依賴：

```text
cmd/watt → internal/orchestrator → internal/pipeline
                                 → internal/runner → internal/proc
                                                   → internal/env
                                 → internal/result
```

- `cmd/watt`：只處理 CLI 參數、旗標與 OS exit code。
- `internal/orchestrator`：選取 pipeline、循序執行、fail-fast、決定結果與 exit code；不得直接啟動 process。
- `internal/pipeline`：YAML 載入、strict decoding、資料模型與靜態驗證；不得探測環境或執行 step。
- `internal/runner`：只執行單一 step、即時透傳及擷取輸出、遮罩已知 env value、判定 step 狀態；不得決定 pipeline 流程。
- `internal/result`：組裝、序列化與寫入結果；視為葉節點，不得反向依賴上層模組或判定業務成敗。
- `internal/env`：host → pipeline → step 的 env 合併、diagnostics 與 redaction context；不得修改 host 環境。
- `internal/proc`：Windows Job Object、process tree 終止與 signal 處理；不得承擔 pipeline 業務語意。

禁止循環依賴、只轉送的抽象層、Phase 2 預留實作，以及對 Taylor 或 Brunel 的編譯期依賴。若參考 Brunel 的 Job Object 做法，必須在本專案獨立實作。

## 實作要求

- 只完成當前任務與驗收條件所需的最小完整變更；不順手重構、改名、格式化或修理無關問題。
- 優先使用 Go 標準庫與既有依賴。新增依賴須有直接需求；未經明確要求不得升級依賴或引入新的主要框架。
- `exec` step 必須直接啟動目標程式，不得經任何 shell 間接執行。
- Shell step 的預設 shell 是 `pwsh`；缺少 PowerShell 7 時必須回報環境不可用，嚴禁 fallback 至 Windows PowerShell 5.1。
- Pipeline 執行必須維持循序與 fail-fast；Phase 1 不得加入平行 step、matrix、timeout 或 `continue-on-error`。
- 每個 step 的 process 必須在任何使用者程式碼執行前綁定 Windows Job Object，建立時即啟用 `JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE` 或等效保證。主行程退出前不得留下任何未確認終止的 descendant process。
- Pipeline 通過載入與靜態驗證後，success、step failure、`environment_unavailable`、handled cancellation 與可組裝的 internal error 都必須盡力產出 partial result。Invalid pipeline 與 `watt check` 不得寫 result。
- Result 中不得包含任何環境變數值。序列化 `resolved_command` 與 `output_tail` 前，依規格 R-8 遮罩 effective environment 中符合條件的已知值；不得把 diagnostic fields 用於 verdict、cache key、result hash 或等價比較。
- 所有不可信輸入只在邊界驗證一次；錯誤須保留原因與可採取行動的上下文，不吞錯、不用未經規格允許的 fallback 或重試掩蓋失敗。

## Pipeline 權限模式

- **Authoring**：只有任務明確要求建立或修改 pipeline 定義時，才可編輯 `watt.yaml` 或其他 pipeline 定義；完成後僅執行 `watt check`，未經人類審核不得將新定義當作正式 validation gate。
- **Execution**：執行既有 `watt run`、讀取 result 並修正專案程式碼時，禁止修改任何 pipeline 定義檔。若 pipeline 無效，回報設定問題，不得為了讓驗證通過而改考卷。

## 測試與驗證

- 新功能與 bug 修復須依可觀察行為補上最小測試；bug 修復必須包含可重現根因的回歸測試。
- 優先執行受影響 package 的針對性測試，再執行 `go test ./...`。涉及 CLI、process、cancellation、env、result 或跨模組流程時，補跑對應 integration／E2E 測試。
- 涉及 Windows Job Object 或 cancellation 時，必須驗證 child 與 grandchild 全部終止、5 秒確認期限、正常結束無孤兒，以及無法確認時回 `EXIT_INTERNAL_ERROR` 而非 `EXIT_CANCELLED`。
- 涉及 result 或 env 時，必須以 canary 驗證 `.watt/result.json` 全文不洩漏已知環境值，並檢查 `exit_code` nullability、partial steps、UTF-8 合法性及每個 stdout/stderr tail 的 8192-byte 上限。
- 涉及 build 或發布產物時，驗證 `windows/amd64`、`CGO_ENABLED=0`、單一 `watt.exe`，且不依賴 Node.js、Python 或額外 runtime。
- 不得宣稱未執行的檢查已通過。受環境限制無法驗證時，明確列出未驗證項目與風險；GitHub Actions 仍是最終權威驗證。
- 若測試專用 TEMP Go module cache 因唯讀屬性無法清除，在確認該路徑確實為本次測試建立且不包含使用者既有資料後，可解除唯讀屬性再執行清理；不得修改其他路徑的權限或使用高風險系統級刪除操作。

## Git 與完成條件

- 開始工作或準備交付前，先執行 `git status --short --branch` 並檢查相關 diff；需要與遠端同步時，先 `git fetch origin`，再確認目前分支與其 upstream 的 ahead／behind 狀態。
- 工作區乾淨且目前分支與 upstream 沒有分歧時，僅使用 `git pull --ff-only` 同步；若有未提交修改、分支分歧或 upstream 不明，停止自動同步並回報，不得自行 stash、merge、rebase、reset 或覆寫檔案。
- 同步後再次檢查 `git status --short --branch` 與最新提交；若同步引入衝突或檢查失敗，保留現場並回報原因，不以跳過檢查的方式繼續。
- 未經使用者明確要求，不得 commit、push、merge、rebase、發布、部署或建立／修改 GitHub 資源。
- 禁止 force push 至 `main`／`master`，禁止破壞性 Git 操作，禁止提交 token、密碼、API key 或其他敏感資料。
- 只有使用者明確要求 closeout 時才重建 `NEXT_ACTION.md`；一般實作完成不得自行改寫工作前線文件。
- 完成時簡潔回報：修改內容、主要行為、已執行驗證、未執行驗證與剩餘風險。
