# Watt — Codex 專案指令

## 專案定位

Watt 是 Windows-first、local-first 的確定性 Pipeline 執行與驗證引擎。Phase 1 以 Go 1.24.x 實作，目標產物為單一 `watt.exe`，static build、`CGO_ENABLED=0`，僅支援 `windows/amd64`。

Watt 只負責載入、驗證與依序執行 pipeline，並產出機器可解析的結果。不得讓核心認知 Taylor、看板、coding agent 或其他上層消費者；整合方向永遠是「消費者 → Watt」。Watt 不提供也不得宣稱提供 sandbox、filesystem isolation、secret store、remote runner 或 GitHub Actions 相容層。

## 開始工作前

1. 先讀 `NEXT_ACTION.md` 確認目前工作前線，再讀任務指定的 GitHub Issue 或其他驗收條件。
2. 依下方 Routing Table 讀取任務相關的 `docs/spec.md` 章節與規則檔，連同直接相關的現有實作、型別、測試與呼叫者；不得依摘要猜測契約。
3. 文件若互相矛盾，以較新且更具體的 `docs/spec.md` 為準；若矛盾會改變公開行為、資料格式或架構方向，停止實作並向使用者確認。
4. 工作區可能含使用者或其他代理的未提交修改。編輯前執行 `git status --short` 並檢查相關 diff；不得覆蓋、還原或順手整理範圍外變更。

## 不可違反規則

- `docs/spec.md` 是功能、架構與驗收標準的唯一權威來源；`DECISIONS.md` 是有效重大決策索引；`PROJECT_BRIEF.md`／`MAZE_PROJECT.md` 僅提供專案背景與工作流定位，不得當作規格依據。
- `docs/spec.md` §7 全部標記為 `[FROZEN]`：Pipeline 資料模型、Result Schema、Exit Code Contract、Process 管理契約不得自行增刪欄位、改名、改值或改變語意；任何變更都必須走 spec revision 並取得使用者確認。
- Phase 1 僅實作 §5.1 Must Have。Should Have 只有在任務明確要求時實作；Could Have、Won't Have 與 §14 列出的 Phase 2 功能不得預作空模組、介面、設定或擴充點。
- 模組依賴只能由上而下（`cmd/watt → orchestrator → pipeline/runner/result`；`runner → proc/env`），禁止循環依賴、只轉送的抽象層、對 Taylor 或 Brunel 的編譯期依賴；細節見 `docs/spec.md` §6。
- Pipeline 權限模式：**Authoring**（只有任務明確要求建立或修改 pipeline 定義時才可編輯 `watt.yaml` 等定義檔，完成後僅執行 `watt check`，未經人類審核不得當作正式 validation gate）；**Execution**（執行既有 `watt run`、讀 result 並修正專案程式碼時，禁止修改任何 pipeline 定義檔；pipeline 無效就回報設定問題，不得為了讓驗證通過而改考卷）。
- Result 中不得包含任何環境變數值；diagnostic fields 不得用於 verdict、cache key、result hash 或等價比較；細節見 `docs/spec.md` §7.2 與 [`docs/agent-rules/result-env.md`](docs/agent-rules/result-env.md)。
- 只完成當前任務與驗收條件所需的最小完整變更；不順手重構、改名、格式化或修理無關問題。優先使用 Go 標準庫與既有依賴；新增依賴或升級／引入新主要框架須有明確需求。
- 不得擴大 Watt 的確定性宣稱：僅 step 順序、fail-fast、env 合併優先序、result schema 與 exit code mapping 屬於保證範圍。公開 CLI、YAML、JSON、exit code 與可觀察行為預設保持相容；實作與規格不一致時修正實作，不得暗改規格來配合程式碼。
- 開始工作或準備交付前先 `git status --short --branch` 並檢查相關 diff；需要同步時先 `git fetch origin`，工作區乾淨且無分歧時僅用 `git pull --ff-only`，其餘情況停止自動同步並回報，不得自行 stash、merge、rebase、reset 或覆寫檔案。未經使用者明確要求，不得 commit、push、merge、rebase、發布、部署或建立／修改 GitHub 資源；禁止 force push 至 `main`／`master`；禁止破壞性 Git 操作；禁止提交 token、密碼、API key 或其他敏感資料。
- 只有使用者明確要求 closeout 時才重建 `NEXT_ACTION.md`；一般實作完成不得自行改寫工作前線文件。
- 不得宣稱未執行的檢查已通過；受環境限制無法驗證時，明確列出未驗證項目與風險。

## Routing Table — 依任務類型讀取

| 任務類型 | 必讀文件／章節 |
|---|---|
| 任何實作任務起手 | `NEXT_ACTION.md`（目前工作前線）＋任務指定的 GitHub Issue |
| Pipeline 資料模型／YAML schema／靜態驗證規則 | `docs/spec.md` §7.1 [FROZEN] |
| Result schema／輸出欄位／partial result | `docs/spec.md` §7.2 [FROZEN] ＋ [`docs/agent-rules/result-env.md`](docs/agent-rules/result-env.md) |
| Exit code 判定 | `docs/spec.md` §7.3 [FROZEN] |
| Process 啟動方式、Windows Job Object、cancellation、shell step | `docs/spec.md` §7.4、§8.2 [FROZEN] ＋ [`docs/agent-rules/process.md`](docs/agent-rules/process.md) |
| env 三層合併、environment diagnostics、已知環境值 redaction | `docs/spec.md` §7.2（R-4／R-8）＋ [`docs/agent-rules/result-env.md`](docs/agent-rules/result-env.md) |
| 撰寫／修改測試、判斷任務驗證是否足夠 | [`docs/agent-rules/testing.md`](docs/agent-rules/testing.md)（測試層次與案例見 `docs/spec.md` §11） |
| 新增／修改 `.github/workflows/**` | [`docs/agent-rules/github-actions.md`](docs/agent-rules/github-actions.md) |
| 架構、模組職責、依賴關係細節 | `docs/spec.md` §6 |
| Repo 現況、檔案位置、目前實作進度 | `REPO_MAP.md` |
| 過去重大決策是否仍有效 | `DECISIONS.md` |
| 專案背景、技術棧、限制摘要 | `PROJECT_BRIEF.md` |

## 完成條件

完成時簡潔回報：修改內容、主要行為、已執行驗證、未執行驗證與剩餘風險。
