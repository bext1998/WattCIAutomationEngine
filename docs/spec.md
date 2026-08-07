# Watt Specification — Phase 1 (MVP)

**版本**：v1.3
**狀態**：Review（Maze 已於 2026-08-08 確認 v1.3 內容）
**日期**：2026-08-08
**適用對象**：實作 AI 代理（Codex / Claude Code / opencode / Cursor）、規格審查者
**技術環境**：Go 1.24.x、static build、zero CGO、Windows-first（amd64）
**前置文件**：[未提供] — Taylor spec 中的 gate 整合章節尚未撰寫，本文件不得反向依賴之

---

## 0. Assumptions（假設彙整）

| # | 假設內容 | 影響範圍 | 確認狀態 |
|---|---|---|---|
| A-1 | Pipeline 定義檔採 YAML，路徑預設 `watt.yaml`（repo root） | CLI、Loader、`pipeline`、`orchestrator`（repo root 解析）、測試 fixture | 待確認 |
| A-2 | Result 輸出預設路徑為 `.watt/result.json`（相對於 repo root） | Result Writer、`orchestrator`（寫入失敗處理）、CLI（讀取）、Taylor 整合、§13 驗收 | 待確認 |
| A-3 | MVP 僅支援 `windows/amd64`；不產出 linux/darwin binary | Build、CI、`proc`、測試、發佈、AC-10 | 待確認 |
| A-4 | 單一 pipeline 檔可含多個具名 pipeline；`watt run` 不帶參數時執行名為 `default` 者 | CLI、`pipeline`（選取邏輯）、`result`（`pipeline` 欄位）、§13 驗收 | 待確認 |
| A-5 | MVP 不含 log 檔持久化，stdout/stderr 即時透傳至終端，result.json 僅存摘要與尾段輸出；寫入 result 前須依 R-4／R-8 redaction 已知環境變數值 | Runner、Result、`env`（redaction）、AC-7 安全測試 | 待確認 |
| A-6 | Watt 為 open-source，授權採 MIT | 專案治理 | 待確認 |
| A-7 | `output_tail` 每個 stream（stdout / stderr）上限為 8192 UTF-8 bytes；保留尾端且序列化後必須為合法 UTF-8 | `runner`、`result`、R-9、AC-11 | 已確認（v1.3 定案，見 R-9） |
| A-8 | Repo root 定義為呼叫 `watt` 指令時之現行工作目錄（cwd），並假設其即為 repository 根目錄 | `pipeline`（cwd 解析）、CLI、A-1、A-2 | 待確認 |
| A-9 | Shell step 未指定 `shell` 欄位時，預設套用 `pwsh`（見 §8.3） | `pipeline` 靜態驗證、`runner` | 待確認 |
| A-10 | Cancellation 時，確認該 step 整棵 descendant process tree 已終止的期限為 5 秒，測量自 Watt 對 Job Object 發出終止訊號起算；逾期未確認即視為需進入 §7.3 的 internal error cleanup flow | `proc`、§4.2、§7.3、AC-6 | 待確認 |

---

## 1. 文件目的與背景

### 1.1 為什麼要做這個

Watt 是 Windows-first、local-first 的**確定性 Pipeline 執行與驗證引擎**。它解決兩個相鄰但不同的問題：

1. **託管 CI 不可用時的執行能力斷層。** GitHub Actions 服務異常、額度耗盡或網路受阻時，repository 與工具鏈其實都在本機，缺的只是一個能照既有定義把 test / build / package 跑完的執行器。
2. **AI 代理產出的可驗證性缺口。** 代理宣稱「已完成」不等於通過驗證。需要一個與代理本身無關的、確定性的判定來源，其結果可被機器直接消費。

這兩個問題共用同一個執行核心，因此不拆成兩個工具。

### 1.2 本文件的角色

本文件是 Watt Phase 1 的主契約文件，定義 CLI 介面、pipeline 資料模型、執行語意、result contract 與 exit code contract。

**消費者與依賴方向（單向，不得反轉）：**

```
人類 ──────────────┐
                   ├──→ Watt CLI ──→ Watt Engine
Coding Agent ──────┤
（透過 Skill）      │
                   │
Taylor（未來）──────┘
（透過外掛程式）
```

**Watt 對上述任何消費者皆無認知。** Watt 不知道 Taylor 存在，不知道看板、卡片或任務狀態，不知道呼叫它的是人還是代理。所有整合皆由消費者端自行完成，方向恆為「消費者 → Watt」。

### 1.3 名詞定義

本文件內下列名詞恆指同一概念，其他措辭僅為口語變體，出現時仍以此表為準：

| 術語 | 定義 |
|---|---|
| pipeline 定義檔 | 使用者撰寫、路徑預設為 `watt.yaml` 的 YAML 檔案；載入後對應 §7.1 的 `PipelineFile` 結構。「pipeline 檔」為同義簡稱 |
| selected pipeline（目標 pipeline） | 依 CLI 參數（或未指定時的 `default`）從 `PipelineFile.Pipelines` 選出、本次要執行的單一 `Pipeline` |
| `environment_unavailable` | Result／Step status 與 exit code 契約中，用以表示「缺少必要 command 或 shell」之唯一正式狀態名稱；文件中「環境問題」「環境缺失」等為其口語描述，判定邏輯一律以此正式值為準 |
| process tree / descendant process | 一個 step 啟動之直接 child process，以及其遞迴衍生之全部子行程；process tree 包含直接 child，並透過 Windows Job Object 綁定並追蹤（見 §7.4） |

---

## 2. 目標（Goals）

- **G-1**：使用者在 Windows 本機執行 `watt run <pipeline>`，可依 pipeline 定義完成 test / build / package，並產出定義中指定的 artifact。artifact 是否與原 CI 產出等價（例如檔案清單、內容 hash），由消費者依專案自訂驗收另行確認，非 Watt 本體職責（Watt 對「原 CI」本身無認知，見 §1.2）。
- **G-2**：pipeline 定義成功載入並通過靜態驗證後，每次 `watt run` 皆盡力產出符合 `[FROZEN]` schema 的 `result.json`；正常寫入成功時必須存在，result 寫入失敗時依 §4.2 回 `EXIT_INTERNAL_ERROR`，可被機器直接解析，無需 parse 人類可讀 log。載入或驗證失敗，以及在完成載入／驗證前取消時不寫 result，例外見 §4.2。
- **G-3**：`watt run` 的 process exit code 可區分至少五種結果類別，供腳本與外部程式分流。
- **G-4**：`watt check` 可在不執行任何 step 的前提下，靜態驗證 pipeline 定義合法性。
- **G-5**：Watt 本體為單一 `watt.exe`，不依賴 Node.js、Python 或任何額外 runtime。
- **G-6**：cancellation 發生時，該 step 所啟動的整棵 process tree 被終止，不留孤兒行程。

---

## 3. 非目標（Non-Goals）

- **NG-1**：不做 GUI。人類介面即 CLI。
- **NG-2**：不設計自製 CI scripting language。Watt 是協調與執行引擎，不是另一種程式語言。
- **NG-3**：不做 GitHub Actions YAML 相容層，不做 Actions Runner 替代品。
- **NG-4**：不提供 Docker / VM / sandbox runner，不提供 filesystem 隔離。**Watt 不宣稱任何 sandbox 能力。**
- **NG-5**：不做 remote runner、分散式 runner、artifact server、secret management platform。
- **NG-6**：不做大型 plugin system。
- **NG-7**：不提供 runtime。專案需要什麼 runtime，由專案本身負責。
- **NG-8**：不感知任何上層消費者（Taylor、看板、任務系統）。
- **NG-9**：不保證外部 command、網路存取、時間或專案本身的執行結果具有確定性（見 §8.2 確定性邊界）。

---

## 4. 使用情境

### 4.1 主要使用情境

**情境 1：託管 CI 中斷時的本機備援**

```
使用者：Maze（人類）
前置狀態：GitHub Actions 服務異常；repository 與工具鏈於本機皆正常
動作：於 repo root 執行 `watt run package`
預期結果：依序執行 Test → Build → Package，成功產出 skills.zip，
          exit code 0，並寫出 .watt/result.json（status: success）
```

**情境 2：Coding Agent 自行驗證（execution 權限）**

```
使用者：Coding Agent（透過 Watt Skill）
前置狀態：Agent 剛完成一次程式修改；pipeline 定義已存在且未被 Agent 修改
動作：執行 `watt run verify --output json`
預期結果：Agent 讀取 result.json，取得 pipeline status 與各 step status；
          失敗時依 failed step 的 stderr 尾段進行修正，再重跑。
          Agent 全程不得修改 pipeline 定義。
```

**情境 3：Pipeline 初次建立（authoring 權限）**

```
使用者：Coding Agent（透過 Watt Skill，authoring 模式）
前置狀態：repository 無 watt.yaml
動作：分析 repo，推導 test / build / package 指令，產生 watt.yaml，執行 `watt check`
預期結果：產出通過靜態驗證的 pipeline 定義草稿，交付人類審核。
          未經人類確認前，該定義不得用於正式 validation。
```

**情境 4：外部程式作為 gate 消費者**

```
使用者：外部程式（例：未來 Taylor 外掛）
前置狀態：Watt 已安裝且 pipeline 已存在
動作：以 subprocess 呼叫 `watt run <pipeline>`，取 exit code 並讀 result.json
預期結果：依 exit code 分流，依 result.json 的 semantic fields 判定通過與否。
          消費者不得依賴 diagnostic fields 做判定。
```

### 4.2 例外情境 / 邊界情況

| 情境 | 預期處理方式 |
|---|---|
| Pipeline 檔不存在 | 回 `EXIT_INVALID_PIPELINE`，訊息含搜尋過的路徑；不寫 result.json |
| Pipeline YAML 語法錯誤 | 回 `EXIT_INVALID_PIPELINE`，訊息含行號；不寫 result.json |
| 指定的 pipeline 名稱不存在 | 回 `EXIT_INVALID_PIPELINE`，列出可用 pipeline 名稱 |
| Step 同時指定 `exec` 與 `run` | 靜態驗證即失敗，`EXIT_INVALID_PIPELINE` |
| Step 兩者皆未指定 | 靜態驗證即失敗，`EXIT_INVALID_PIPELINE` |
| `exec` 指定的執行檔不在 PATH | 回 `EXIT_ENVIRONMENT_UNAVAILABLE`，明確指出缺少哪個 command；寫出 partial result |
| `shell: pwsh` 但系統無 PowerShell 7 | 回 `EXIT_ENVIRONMENT_UNAVAILABLE`，明確回報缺少 `pwsh`。**嚴禁 fallback 至 Windows PowerShell 5.1** |
| Step 非零 exit code | fail-fast：中止 pipeline，回 `EXIT_STEP_FAILED`，寫出 partial result |
| `cwd` 指向不存在的目錄 | 該 step 的執行嘗試標記 `failed`，回 `EXIT_STEP_FAILED`；此時 step 可沒有 OS exit code |
| 使用者於 step 執行中按下 Ctrl+C，且該 step 整棵 descendant process tree 於 5 秒內確認終止 | 回 `EXIT_CANCELLED`，盡力寫出 partial result（見 §7.3 不變式） |
| 使用者於 step 執行中按下 Ctrl+C，但於期限內無法確認 process tree 已全部終止 | 進入 internal error cleanup flow；完成 P-3 的安全清理後回 `EXIT_INTERNAL_ERROR`，盡力寫出 `status: internal_error` 的 partial result；**嚴禁在仍可能存在孤兒行程時退出或謊報 `EXIT_CANCELLED`**（見 §7.3 不變式） |
| 使用者於 step 之間（前一 step 已結束、下一 step 尚未啟動）按下 Ctrl+C | 不啟動下一個 step，回 `EXIT_CANCELLED`，寫出目前已執行 step 的 partial result |
| 使用者於任何 step 啟動前（載入、驗證、env 合併、cwd 解析階段）按下 Ctrl+C | 若 pipeline 已成功載入並通過驗證，回 `EXIT_CANCELLED` 並寫出 `status: cancelled`、空 `steps` 的 result；若尚未完成載入或驗證，回 `EXIT_CANCELLED` 且不寫 result.json |
| `watt check --env` 偵測到缺少必要 command 或所需 shell | 回 `EXIT_ENVIRONMENT_UNAVAILABLE`，訊息列出缺少項目；`check` 不產生 result.json（見 F-3、F-4） |
| result.json 寫入失敗（例如磁碟空間不足、權限不足） | 回 `EXIT_INTERNAL_ERROR`；若已啟用 `--output json`，仍盡力將 result 內容輸出至 stdout 供人工搶救 |
| Step 輸出超大量 stdout | 即時透傳終端；result.json 僅保留尾段（見 §7.2 `output_tail`） |

**補充說明（result.json 的新舊區分）**：載入或靜態驗證失敗時 Watt 不寫出 result.json，且不觸碰檔案系統中既有的 `.watt/result.json`（不覆寫、不刪除）。消費者若在 `EXIT_INVALID_PIPELINE` 之後讀取到 result.json，該檔案可能是先前一次成功執行留下的舊結果，而非本次執行的產物。消費者應以 process exit code 作為判定第一依據：`EXIT_SUCCESS`、`EXIT_STEP_FAILED`、`EXIT_ENVIRONMENT_UNAVAILABLE`、`EXIT_CANCELLED` 及可組裝 result 的 `EXIT_INTERNAL_ERROR`，其 result.json（若存在）才可能是本次執行所寫出；`EXIT_INVALID_PIPELINE` 恆不保證，result 寫入失敗時亦不保證檔案存在。

---

## 5. 功能規格

### 5.1 Must Have（Phase 1 必須完成）

- **F-1**：`watt run` — 執行 `default` pipeline。
- **F-2**：`watt run <pipeline>` — 執行指定 pipeline。
- **F-3**：`watt check` — 僅做靜態 pipeline 驗證（不執行 step、不探測環境、無副作用、可離線）。
- **F-4**：`watt check --env` — 靜態驗證 + 環境探測（檢查 `exec` 目標與所需 shell 是否可解析）。
- **F-5**：Exec Step — 不經 shell 直接啟動外部程式（`exec` + `args`）。
- **F-6**：Shell Step — 經指定 shell 執行 script（`shell` + `run`）。MVP 支援 `pwsh` 與 `cmd`。
- **F-7**：Step 級 `cwd` 與 `env` 支援。
- **F-8**：Environment inheritance — host env → pipeline env override → step env override。
- **F-9**：Fail-fast 執行語意。
- **F-10**：`result.json` 產出，符合 §7.2 `[FROZEN]` schema。
- **F-11**：Exit code contract，符合 §7.3 `[FROZEN]`。
- **F-12**：Windows Job Object process tree 管理與 cancellation。
- **F-13**：stdout / stderr 即時透傳至呼叫端終端。

### 5.2 Should Have（不擋 Phase 1 驗收）

- **F-14**：`--output json` 旗標，於可組裝 result 的執行結束後將完整 result JSON 寫至 stdout（供 pipe 消費）。**啟用此旗標時，各 step 的即時輸出改導向呼叫端 stderr**（不再依 F-13 預設寫入 stdout），確保呼叫端 stdout 全程只承載最終唯一的 JSON，pipe 消費者可直接解析；若 invalid pipeline 等情況無 result 可組裝，錯誤訊息改寫至 stderr，stdout 不保證有 JSON。未啟用 `--output json` 時維持 F-13 預設行為：step 輸出即時透傳至 stdout/stderr。
- **F-15**：`shell: bash` 支援（由使用者環境提供 Git Bash / WSL，Watt 不附帶）。
- **F-16**：`watt run --dry-run` — 印出將執行的 step 序列與解析後的 command，不實際執行。

### 5.3 Could Have（Phase 1 不實作）

- **F-17**：Step 級 `continue-on-error`。
- **F-18**：Step 級 timeout。
- **F-19**：Pipeline 間相依（`needs`）。
- **F-20**：Log 檔持久化與 log 輪替。
- **F-21**：Strict env 模式（no-inherit / allowlist）。

### 5.4 Won't Have（明確排除）

- **F-22**：平行 step 執行 — 破壞確定性 ordering 與 fail-fast 語意的可預測性，Phase 1 不做。
- **F-23**：Matrix build — 屬託管 CI 特徵，非本地備援核心需求。
- **F-24**：Secret store — 見 NG-5。

---

## 6. 系統架構

### 6.1 模組總覽

```
[Entry Layer]
  └── cmd/watt              cobra CLI：run / check
       │
[Application Layer]
  └── internal/orchestrator 決定執行哪個 pipeline、串接下列模組、決定 exit code
       │
[Domain Layer]
  ├── internal/pipeline     Pipeline / Step 資料模型 + 靜態驗證
  ├── internal/runner       單一 step 的執行、輸出擷取、已知 env value redaction、狀態判定
  ├── internal/result       Result 組裝、序列化與落檔
  └── internal/env          env 三層合併、diagnostics 蒐集與 redaction context
       │
[Platform Layer]
  └── internal/proc         Windows Job Object、process tree 終止、signal 處理
```

### 6.2 模組職責

| 模組 | 職責 | 不負責 |
|---|---|---|
| `cmd/watt` | 參數解析、旗標、將 exit code 交還 OS | 執行邏輯、驗證邏輯 |
| `orchestrator` | pipeline 選取、step 迭代、fail-fast 決策、exit code 決定、觸發 result 寫出 | 直接啟動 process |
| `pipeline` | YAML 載入、資料模型、靜態驗證規則 | 環境探測、執行 |
| `runner` | 單一 step 執行、stdout/stderr 擷取、已知 env value redaction、exit code 取得、step status 判定 | 決定下一個 step 是否執行 |
| `result` | Result 結構組裝、schema 版本控制、序列化與落檔 | 判定成功或失敗 |
| `env` | 三層 env 合併、產出經 redaction 的 environment diagnostics、提供 result redaction context | 修改 host 環境 |
| `proc` | Job Object 建立與綁定、descendant 終止、Ctrl+C handler | 業務語意 |

### 6.3 依賴關係

```
cmd/watt → orchestrator → pipeline
                       → runner → proc
                               → env
                       → result
```

**依賴方向規則**：僅允許由上而下。`runner` 不得認知 pipeline 整體流程；`pipeline` 不得 import `runner`；`result` 為葉節點，不得反向 import 任何上層模組。禁止任何循環依賴。

---

## 7. 介面定義 [FROZEN]

> 本節所有內容標記為 `[FROZEN]`。實作代理不得自行增刪欄位、改名或改變語意。修改需走 spec revision。

### 7.1 Pipeline 資料模型 [FROZEN]

```go
type PipelineFile struct {
    Version   int                 `yaml:"version"`   // 必填，MVP 固定為 1
    Env       map[string]string   `yaml:"env"`       // 選填：pipeline 級 env override
    Pipelines map[string]Pipeline `yaml:"pipelines"` // 必填，至少一組
}

type Pipeline struct {
    Steps []Step `yaml:"steps"` // 必填，至少一個 step
}

type Step struct {
    Name  string            `yaml:"name"`  // 必填：唯一識別，同一 pipeline 內不得重複
    Exec  string            `yaml:"exec"`  // 與 Run 互斥；指定執行檔
    Args  []string          `yaml:"args"`  // 選填：僅在 Exec 模式有效
    Run   string            `yaml:"run"`   // 與 Exec 互斥；shell script 內容
    Shell string            `yaml:"shell"` // 僅在 Run 模式有效；MVP Must Have：pwsh | cmd；bash 屬 F-15 Should Have，未指定時預設 pwsh（見 §8.3、A-9）
    Cwd   string            `yaml:"cwd"`   // 選填：相對於 repo root，預設為 repo root
    Env   map[string]string `yaml:"env"`   // 選填：step 級 env override（最高優先）
}
```

**互斥不變式**：`Exec` 與 `Run` 必須恰有一個為非空。違反者於靜態驗證階段失敗。

**靜態驗證規則彙整 [FROZEN]**：

- `version` 必須為整數 `1`；其他值視為驗證失敗。
- `pipelines` 至少須有一組具名 pipeline；pipeline 名稱不得為空字串；每個 `Pipeline` 的 `steps` 至少一個。`watt run` 未指定名稱時若不存在 `default`，視為 pipeline 名稱不存在並回 `EXIT_INVALID_PIPELINE`。
- Step `name` 於同一 pipeline 內不得重複，且不得為空白字串。
- `Exec` 與 `Run` 必須恰有一個非空（去除前後空白後仍有內容；見上方互斥不變式）。
- `Args` 僅在 `Exec` 模式（`Exec` 非空）下允許包含項目；`Run` 模式下提供非空 `Args` 視為欄位誤用，驗證失敗；空 list 等同未指定。
- `Shell` 僅在 `Run` 模式（`Run` 非空）下有意義；`Exec` 模式下提供非空 `Shell` 視為欄位誤用，驗證失敗；空字串等同未指定。
- `Run` 模式下 `Shell` 未指定時預設為 `pwsh`（見 A-9、§8.3）；MVP 僅接受 `pwsh` 與 `cmd`，指定 `bash` 於 F-15 實作前視為驗證失敗並提示尚未支援。
- 未知欄位一律視為驗證失敗（strict decoding），避免拼寫錯誤導致欄位靜默失效。

### 7.2 Result Schema [FROZEN]

Result 欄位在概念上分為兩類，此分類本身即為契約的一部分：

- **Semantic fields** — 可用於 verdict 判定、cache key、result hash、等價比較。
- **Diagnostic fields** — 僅供人類除錯與事後分析。**任何消費者不得將其納入判定、快取鍵或等價比較。**

```jsonc
{
  // ---------- semantic ----------
  "schema_version": 1,              // 單調遞增 integer，非 semver
  "pipeline": "package",
  "status": "success",              // success | failed | environment_unavailable | cancelled | internal_error
  "steps": [
    {
      // ---------- semantic ----------
      "name": "Test",
      "status": "success",          // success | failed | cancelled | environment_unavailable
      "exit_code": 0,                // 整數；step 未能取得 OS exit code 時為 null（見 R-7）

      // ---------- diagnostic ----------
      "started_at": "2026-08-07T10:00:00+08:00",
      "duration_ms": 12043,
      "resolved_command": "go test ./...",
      "output_tail": {
        "stdout": "...",            // 尾段，上限 8192 UTF-8 bytes（見 R-9）
        "stderr": "..."
      }
    }
  ],

  // ---------- diagnostic ----------
  "started_at": "2026-08-07T10:00:00+08:00",
  "duration_ms": 45120,
  "watt_version": "0.1.0",
  "environment": {
    "os": "windows",
    "arch": "amd64",
    "shell_available": { "pwsh": "7.4.6", "cmd": true, "bash": false },
    "resolved_tools": { "go": "C:\\Program Files\\Go\\bin\\go.exe" },
    "env_var_names": ["PATH", "GOPATH", "GITHUB_TOKEN"]
  }
}
```

**Result 不變式 [FROZEN]：**

- **R-1**：`duration_ms`、`started_at`、`output_tail`、`resolved_command`、`environment`、`watt_version` 皆為 diagnostic，不參與 verdict、cache key、result hash 或等價比較。
- **R-2**：`steps` 僅包含 orchestrator 已選取、且 runner 已開始嘗試執行的 step；這包括 cwd／command／shell 解析或 process 啟動失敗的嘗試。尚未嘗試執行者不列入 MVP result（不以 `skipped` 佔位）。
- **R-3**：pipeline 已成功載入並通過驗證後，result.json 在 success、step failure、`environment_unavailable`、handled cancellation 及可組裝的 internal error 情況下皆須盡最大努力寫出。**嚴禁僅於 success path 寫檔。** 載入／驗證失敗及 `watt check` 不產生 result.json。
- **R-4**：`environment` 僅得包含安全 diagnostics。**嚴禁輸出任何環境變數的值。** 允許輸出：variable name 清單、OS/arch、shell 版本、解析到的 tool path。禁止輸出：任何 variable value，即使經過 hash。已知的 effective environment values 亦不得出現在 `resolved_command` 或 `output_tail` 的序列化結果中。此限制之目的為避免 API key、GitHub token、registry credential 經由 result artifact 或 Agent context 外洩；即時透傳至終端的原始輸出不在此 redaction 契約內。
- **R-5**：`schema_version` 為單調遞增 integer。消費者僅得做相等或大小比較，不得解析版本語意。
- **R-6**：新增 diagnostic 欄位不遞增 `schema_version`；新增 semantic 欄位、移除欄位、改名或改變既有欄位語意則必須遞增。本規則自 Phase 1 首次實作並產出 result.json 起生效；規格本身在正式實作前的草稿修訂（包含本次新增 `internal_error` 狀態值）不視為對已發行 schema 的變更，不因此遞增 `schema_version`。
- **R-7**：`exit_code` 為 semantic 欄位；當 step 因 process 從未啟動（包含 `environment_unavailable`，以及 `cwd` 解析失敗等其他導致 process 無法啟動的 `failed` 情況）或因 cancellation 而未能取得作業系統 exit code 時，`exit_code` 為 `null`。僅有實際啟動並已終止之 process 才具備非 null 的整數 `exit_code`；`status` 為 `failed` 不保證 `exit_code` 非 null，消費者不得假設兩者恆同時成立。若 cancellation 的 process tree 無法在期限內確認終止，且仍能組裝 partial result，top-level `status` 必須為 `internal_error`；受影響的該 step 本身仍標記 `status: "cancelled"`（`exit_code` 依前述規則為 `null`），不得將 top-level 結果當作 handled cancellation。
- **R-8**：`output_tail`、`resolved_command` 為 diagnostic 欄位。序列化前必須遮罩 effective environment 中長度達 8 個字元（含）以上的已知非空 value；短於此門檻的 value（例如 `"1"`、`"8"` 等常見旗標或數字）不進行逐字比對遮罩，避免對 step 自身輸出的無關數字、路徑片段造成大量誤傷。遮罩不延伸至即時透傳的終端輸出，也不保證識別 pipeline 內未出現在 effective environment 的秘密常值，或短於門檻的敏感值。遮罩後的結果仍不得被消費者納入 verdict、cache key、result hash 或等價比較。
- **R-9**：`output_tail.stdout` 與 `output_tail.stderr` 必須存在且為字串；每個 stream 上限為 8192 個 UTF-8 bytes，超過時保留尾端，並確保截斷後仍為合法 UTF-8。無輸出時使用空字串。

**Partial Result 欄位規則**：

- `status: "environment_unavailable"` 的 step：`exit_code` 為 `null`；`resolved_command` 若指令本身仍可解析則填入，否則為空字串；`output_tail` 為 `{ "stdout": "", "stderr": "" }`（尚未產生任何輸出）。
- `status: "failed"` 但 process 從未啟動的 step（例如 `cwd` 指向不存在的目錄）：`exit_code` 為 `null`；`resolved_command` 若指令本身仍可解析則填入，否則為空字串；`output_tail` 為 `{ "stdout": "", "stderr": "" }`。`status: "failed"` 的其餘情況（process 已啟動並回傳非零 exit code）`exit_code` 為非 null 整數。
- `status: "cancelled"` 的 step：`exit_code` 為 `null`，除非底層 process 已回報終止碼，此時允許填入該值供除錯；`output_tail` 保留取消前已擷取的內容。此規則同樣適用於 top-level `status` 為 `internal_error`、但該 step 因取消而受影響的情況（見 R-7）。
- top-level `status: "internal_error"`：僅適用於 pipeline 已成功載入並通過驗證、且 Watt 能在錯誤發生後組裝 result 的情況；若 result 寫入本身失敗，檔案可能不存在，但 `--output json` 應盡力輸出該結果。

### 7.3 Exit Code Contract [FROZEN]

| 常數 | 值 | 觸發條件 | 消費者處理建議 |
|---|---|---|---|
| `EXIT_SUCCESS` | 0 | 所有 step 皆成功 | 判定通過 |
| `EXIT_STEP_FAILED` | 1 | 某 step 回非零 exit code，或 step 的執行嘗試因 cwd 等 step 層錯誤而標記 `failed` | 判定不通過；讀 result 找出 failed step |
| `EXIT_INVALID_PIPELINE` | 2 | `watt run`／`watt check` 的 pipeline 檔缺失、語法錯誤、驗證失敗、名稱不存在 | 設定問題，非程式碼問題；不應重試 |
| `EXIT_ENVIRONMENT_UNAVAILABLE` | 3 | `watt run` 缺少必要 command 或 shell；或 `watt check --env` 偵測到同類缺失 | 環境問題，非程式碼問題；不應判定為驗證失敗 |
| `EXIT_CANCELLED` | 4 | 收到中斷訊號，且無 active step，或 active step 的整棵 descendant process tree 已確認終止（見下方不變式） | 結果不具判定意義；不得視為失敗 |
| `EXIT_INTERNAL_ERROR` | 5 | Watt 自身異常，包含但不限於：(a) cancellation cleanup 經重試仍無法確認 process tree 已全部終止、(b) result.json 寫入失敗、(c) 未預期的 panic 經 recover 攔截 | 回報 bug；不得視為驗證失敗 |

**不變式**：`EXIT_CANCELLED` 僅得在**無 active step，或 active step 的整棵 descendant process tree 於確認期限內確認終止**後回傳；此確認期限見 A-10（現訂為 5 秒，測量自 Watt 對 Job Object 發出終止訊號起算）。所有取消流程仍須完成 P-3 的清理要求；若逾此期限仍無法確認終止，先進入 internal error cleanup flow，完成安全清理（包括必要時啟用平台保證的 Job Object close-on-kill 機制）後才回傳 `EXIT_INTERNAL_ERROR`，不得在仍可能存在孤兒行程時退出。§4.2、AC-6 所稱之「5 秒」皆指同一個 A-10 期限，不另行定義。

對 `watt check` 而言，靜態驗證成功回 `EXIT_SUCCESS`；驗證失敗回 `EXIT_INVALID_PIPELINE`。`watt check --env` 若額外偵測到必要 command 或 shell 缺失，回 `EXIT_ENVIRONMENT_UNAVAILABLE`。

### 7.4 Process 管理契約 [FROZEN]

- **P-1**：每個 step 啟動的 process 必須綁定至 Windows Job Object，且綁定須發生於 process 開始執行任何使用者程式碼之前（例如以 suspended 狀態建立後綁定、再 resume），避免 process 建立與加入 Job Object 之間出現可逃逸的競態視窗。Job Object 建立當下必須設定 `JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE`（或等效 limit），確保 handle 關閉時由作業系統強制終止其內全部 process；P-3 所稱「平台保證的 close-on-kill 機制」即依賴此處預先設定的 limit，事後才設定無效。
- **P-2**：cancellation 時，終止該 Job Object 內全部 process tree（含直接 child 與其 descendants）。
- **P-3**：無論 step 正常結束、失敗或被取消，Watt 在關閉對應 Job Object 前必須確認其內已無存活 process；若期限內無法確認，必須繼續終止／等待並保留 Job Object 與監督流程，不得靜默關閉或退出。若採用平台保證的 Job Object close-on-kill 機制作為最後清理手段，必須先啟用該機制並成功關閉 Job Object，才視為完成安全清理；若既無法確認無存活 process、也無法建立該平台保證，Watt 不得退出。安全清理完成後若原流程屬 internal error，回傳 `EXIT_INTERNAL_ERROR`。Watt 主行程退出前，所有已建立的 Job Object 皆須完成此確認或平台保證的終止並關閉，不得留下孤兒行程。
- **P-4**：`exec` 模式禁止經由任何 shell 間接啟動。

---

## 8. 運作原理

### 8.1 主要流程

```
載入 pipeline 檔 → 解析失敗則 EXIT_INVALID_PIPELINE（不寫 result）
      ↓
靜態驗證（依 §7.1 全部規則）→ 失敗則 EXIT_INVALID_PIPELINE（不寫 result）
      ↓
選取目標 pipeline
      ↓
┌─→ 取下一個 step
│         ↓
│   合併 env（host → pipeline → step）
│         ↓
│   解析 cwd、解析 command
│         ↓
│   建立 Job Object → 啟動 process
│         ↓
│   即時透傳 stdout / stderr，同時保留尾段
│         ↓
│   取得 exit code → 判定 step status
│         ↓
│   成功 ─────────────┘（若尚有 step）
│   失敗 / environment_unavailable / 取消 → 中止迭代
      ↓
組裝 result（含已開始嘗試的 step）→ 寫出 result.json
      ↓
決定並回傳 exit code
```

### 8.2 確定性邊界 [FROZEN]

Watt 之「確定性」為**有界宣稱**，實作與文件皆不得誇大：

**Watt 保證 deterministic：**
- Step 執行順序
- Fail-fast 錯誤傳播語意
- Env 三層合併的優先順序
- Result schema 結構與欄位語意
- Exit code 對應關係

**Watt 明確不保證 deterministic：**
- 外部 command 的執行結果
- 網路存取結果
- 時間相關行為
- 專案本身的測試或建置結果
- 執行耗時

### 8.3 關鍵設計決策

| 決策 | 選擇 | 理由 |
|---|---|---|
| 預設 shell | `pwsh`，缺少時明確失敗 | Unicode / JSON / 檔案處理完整；與 Actions Windows runner 一致。靜默 fallback 至 5.1 會造成行為不一致，屬確定性破口 |
| `exec` 與 `shell` 二分 | `exec` 為一等公民，不依賴 shell | 使 Watt 本體維持零外部 runtime 依賴；`pwsh` 僅為 shell step 的選配外部依賴 |
| `watt check` 語意 | 預設純靜態，環境探測走 `--env` | Gate 場景只需靜態驗證，速度差異顯著；且靜態驗證須可離線、無副作用 |
| Env 策略 | 保留 host 繼承，不做白名單 | Windows 工具鏈對 PATH、TEMP、SDK 變數有深度隱性依賴，MVP 上嚴格白名單會使多數 pipeline 無法執行 |
| Env 記錄策略 | 僅記錄 variable name，不記錄任何 value | 完整 env dump 極易挾帶 secret 進入 artifact、git 或 Agent context |
| Partial result | 對已載入並通過驗證的 `watt run` 一律盡力寫出；`watt check` 與 invalid pipeline 不寫 | 機器消費者最需要結果的時機正是失敗時 |
| Duration 分類 | Diagnostic | 非決定性資料若進入 cache key 或等價比較會造成誤判 |
| Agent 權限 | authoring / execution 分離 | 正式 validation 時若代理可改 pipeline，等同受測者可改考卷 |

### 8.4 Agent 權限語意 [FROZEN]

| 模式 | 允許 | 禁止 |
|---|---|---|
| **Authoring** | 分析 repo、產生或修改 pipeline 定義、執行 `watt check` | 將未經人類審核的定義用於正式 validation |
| **Execution** | 執行 `watt run`、讀取 result.json、依 stderr 修正**專案程式碼** | 修改任何 pipeline 定義檔 |

**此為 Harness / Skill 層級之 policy，非 Watt 提供之技術性強制。** Watt 不具備 filesystem sandbox，spec 與 Skill 文件皆不得暗示其具備。

---

## 9. 替代方案

### 方案 A：單一引擎，雙消費者（推薦）

**做法**：Watt 只有一套執行核心與一組輸出契約，同時服務人類 CLI 與機器消費者。JSON result 列為 MVP MUST。

**優點**：
- 單一事實來源，人類看到的與機器判定的是同一次執行
- Taylor 整合時無需修改 Watt 核心，只需在消費端寫外掛
- 契約層在 MVP 即凍結，後期成本趨近於零

**缺點 / 限制**：
- MVP 範圍略大於純備援 CI
- Result schema 一旦凍結，早期設計失誤的修正成本較高

### 方案 B：先做備援 CI，Taylor 整合時再補 machine output

**做法**：MVP 只求人類可用，JSON 輸出降為 Should，待實際整合時再定義。

**優點**：
- MVP 範圍最小，可最快跑通驗收情境
- 契約在真實整合需求出現後才定，較不易設計錯

**缺點 / 限制**：
- Result contract 是典型「後加會痛、前加免費」的項目
- 消費者極可能先寫出 parse 人類 log 的權宜方案，此類程式碼難以移除
- 與 Watt 作為確定性驗證來源的定位相矛盾

### 推薦方向

選擇**方案 A**。決定性因素在於：契約層（result schema、exit code、確定性邊界、權限界線）的增量實作成本在 MVP 階段極低，而延後導入的代價是消費端會長出依賴人類可讀輸出的權宜實作。此類實作一旦存在便難以移除，且會反向侵蝕 Watt 的定位。

---

## 10. 編程範例

### 10.1 Pipeline 定義基本用法

```yaml
version: 1

env:
  CGO_ENABLED: "0"

pipelines:
  package:
    steps:
      - name: Test
        exec: go
        args: ["test", "./..."]

      - name: Build
        exec: go
        args: ["build", "-o", "dist/watt.exe", "./cmd/watt"]

      - name: Package
        shell: pwsh
        run: |
          Remove-Item ./dist/skills -Recurse -Force -ErrorAction SilentlyContinue
          Compress-Archive ./skills/* ./dist/skills.zip -Force
```

### 10.2 消費者正確用法 vs 常見錯誤

```go
// ✅ 正確：先依 exit code 分流，再讀 semantic fields
switch exitCode {
case 0:
    return VerdictPass
case 1:
    return VerdictFail
case 2:
    return VerdictInvalidPipeline
case 3, 4, 5:
    return VerdictInconclusive // 環境問題 / 取消 / 內部錯誤皆非驗證失敗
}

// ❌ 錯誤：把非零一律當驗證失敗
if exitCode != 0 { return VerdictFail } // 缺 pwsh 會被誤判為程式碼有問題

// ❌ 錯誤：把 diagnostic fields 納入 cache key
key := hash(result.Pipeline, result.Status, result.DurationMs) // duration 非決定性

// ❌ 錯誤：parse 人類可讀輸出取代 result.json
if strings.Contains(stdout, "FAIL") { ... }
```

---

## 11. 測試策略

### 11.1 測試層次

| 層次 | 範圍 | 工具 | 覆蓋目標 |
|---|---|---|---|
| Unit | 單一函式 | Go `testing` | 靜態驗證規則、env 合併優先序、exit code 對應、redaction 邏輯 |
| Integration | 跨模組 | Go `testing` | orchestrator + runner + proc 的 fail-fast 與 cancellation |
| E2E | 完整 CLI | Go `testing` + testdata fixtures | §13 全部驗收情境 |

### 11.2 關鍵測試案例

- `TestValidate_ExecAndRunBothSet` — 互斥不變式
- `TestValidate_NeitherExecNorRun` — 互斥不變式
- `TestValidate_DuplicateStepName` — 名稱唯一性
- `TestEnvMerge_StepOverridesPipelineOverridesHost` — 三層優先序
- `TestEnvDiagnostics_NoValuesLeaked` — **result 中不得出現任何 env value**（以 canary value 植入 host env 驗證）
- `TestRun_FailFastStopsAtFailedStep` — 後續 step 不得執行
- `TestResult_PartialOnFailure` — steps 僅含已開始嘗試執行者
- `TestResult_WrittenOnCancellation` — 取消時仍寫檔
- `TestCheck_NoSideEffects` — `watt check` 不啟動任何 process
- `TestMissingPwsh_NoFallbackTo51` — 明確失敗，不得 fallback
- `TestCancel_KillsDescendantProcesses` — 啟動 grandchild process，驗證 cancel 後全數消失
- `TestSchemaVersion_IsInteger` — schema 契約
- `TestRun_DefaultPipelineSelection` — 不帶參數時執行 `default` pipeline（F-1）
- `TestRun_NamedPipelineSelection` — 指定 pipeline 名稱時正確選取（F-2）
- `TestCheckEnv_DetectsMissingShell` — `watt check --env` 偵測缺少 `pwsh`，回 `EXIT_ENVIRONMENT_UNAVAILABLE`（F-4）
- `TestExecStep_NoShellIndirection` — 驗證 `exec` 模式未經任何 shell 間接啟動（P-4）
- `TestShellStep_CmdSupport` — `cmd` shell 執行成功（F-6）
- `TestStep_CwdResolvedRelativeToRepoRoot` — `cwd` 相對路徑正確解析（F-7）
- `TestOutputPassthrough_RealtimeStdoutStderr` — step 輸出即時透傳至呼叫端終端（F-13）
- `TestOutputJSON_StdoutCarriesOnlyFinalResult` — 啟用 `--output json` 時 stdout 僅含最終 JSON，step 輸出改走 stderr（F-14）
- `TestExitCode_InternalErrorOnUnconfirmedCancellation` — 無法確認 process tree 終止時回 `EXIT_INTERNAL_ERROR`，不得謊報 `EXIT_CANCELLED`
- `TestProc_NoOrphansOnNormalCompletion` — 正常結束（非取消路徑）亦無孤兒行程（P-3）
- `TestResult_ExitCodeNullOnEnvironmentUnavailable` — `environment_unavailable` step 的 `exit_code` 為 `null`（R-7）
- `TestResult_ExitCodeNullOnCwdFailure` — `cwd` 指向不存在目錄時，該 `failed` step 的 `exit_code` 為 `null`，`output_tail` 為空字串（R-7、Partial Result 欄位規則）
- `TestResult_OutputTail_RedactsKnownEnvValues` — step 將 effective environment 的 canary value 印至 stdout/stderr 時，result 的 `output_tail`／`resolved_command` 不含該 value；即時終端透傳不在此測試範圍（R-4、R-8）
- `TestResult_InternalErrorStatus` — 可組裝 partial result 的 internal error 使用 top-level `status: "internal_error"`，不得被消費者視為驗證失敗（R-3、R-7）

### 11.3 不測試的範圍

- 外部工具鏈本身的正確性（go、npm 等）
- PowerShell 自身行為
- 非 Windows 平台（MVP 不支援）

---

## 12. 風險與限制

| 風險 | 嚴重度 | 緩解方式 |
|---|---|---|
| Result schema 早期設計失誤，凍結後修正成本高 | 高 | `schema_version` 為 integer，允許破壞性遞增；`environment` 區塊為 diagnostic，變更不影響消費者判定邏輯 |
| Job Object 處理不完整，cancellation 留孤兒行程 | 高 | 沿用 Brunel 既有 Job Object 實作經驗；以 grandchild process 測試驗證；無法確認終止時回 `EXIT_INTERNAL_ERROR` 而非謊報 cancelled |
| Env value 意外進入 result artifact | 高 | R-4／R-8 明文禁止；以子程序輸出 canary test 驗證；即時終端輸出仍由呼叫端自行保護 |
| Agent 在 execution 模式下擅自修改 pipeline | 中 | Skill 文件明文禁止；建議消費端於執行前後比對 pipeline 檔 hash（屬消費者責任，非 Watt 功能） |
| `pwsh` 非 Windows 內建，構成外部依賴 | 中 | `exec` 為一等公民且零 shell 依賴；`watt check --env` 提前暴露缺失；文件明示 shell step 為選配能力 |
| 消費者誤用 diagnostic fields | 中 | Schema 內以 semantic / diagnostic 分類註記；§10.2 提供反例；R-1 明文禁止 |
| MVP 範圍蔓延至 Phase 2 | 中 | §14 Phase 邊界聲明；§5.4 Won't Have |
| 大量 stdout 造成記憶體壓力 | 低 | 僅保留尾段，即時透傳不緩衝全文 |

---

## 13. 驗收標準

| # | 驗收項目 | 測量方式 | 通過標準 |
|---|---|---|---|
| AC-1 | 備援 CI 核心情境 | 於離線環境執行 `watt run package` | 產出 skills.zip，exit code 0，result.status 為 success |
| AC-2 | Fail-fast | 令第二個 step 回非零 | exit code 1；result.steps 長度為 2；第三個 step 不出現在 result |
| AC-3 | Command not found | `exec` 指向不存在之執行檔 | exit code 3；錯誤訊息含缺少的 command 名稱；無 panic |
| AC-4 | Invalid pipeline | 提供語法錯誤之 YAML | exit code 2；訊息含行號；未寫出 result.json；既有 result.json 不被覆寫或刪除 |
| AC-5 | Missing shell | 於無 `pwsh` 環境執行 shell step | exit code 3；訊息指出缺少 pwsh；**未執行任何 PowerShell 5.1 指令** |
| AC-6 | Cancellation | 執行長時間 step 期間送出 Ctrl+C | process tree 於 5 秒內全數消失且完成 P-3；成功確認時 exit code 4、result.status 為 cancelled；若需 internal error，完成安全清理後 exit code 5、result.status 為 internal_error |
| AC-7 | Secret 不外洩 | 於 host env 植入 canary 值，並讓 step 將該值印至 stdout/stderr | result.json 全文不含該 canary 值；即時終端透傳不在此 AC 的保證範圍 |
| AC-8 | Check 無副作用 | 對含破壞性 step 的 pipeline 執行 `watt check` | exit code 0；檔案系統無任何變更；無 process 被啟動 |
| AC-9 | 零 runtime 依賴 | 於未安裝 Node.js / Python 的環境執行 `watt --version` | 正常輸出 |
| AC-10 | 單一執行檔 | 檢視 build 產物 | 單一 `watt.exe`，static build，無 CGO |
| AC-11 | Result schema 一致性 | 對 success、failed、environment_unavailable、cancelled 四種正常結果各執行一次 | 四次皆產出符合 §7.2 的 result.json（AC-4 除外）；`output_tail` 欄位與 `exit_code` nullability 符合 R-7／R-9 |
| AC-12 | Internal error result | 觸發可組裝 result 的 internal error 路徑 | exit code 5；result.status 為 internal_error；消費者不得判定為驗證失敗 |

---

## 14. Phase 邊界聲明

> **Phase 1 不得實作以下 Phase 2 功能**，即使技術上可行、即使對話中曾被提及：

- GUI（任何形式）
- Taylor 外掛或任何 Taylor 專用整合程式碼
- 平行 step 執行
- Docker / VM / sandbox runner
- Remote / 分散式 runner
- Artifact server、secret management
- Plugin system
- GitHub Actions YAML 相容層
- Strict env 模式
- Log 檔持久化

**預留擴充點說明**：`result` 模組的 schema 版本機制與 `runner` 的 step 介面已預留擴充空間，但 Phase 1 不得填入任何上述功能之實作邏輯，亦不得為其預先建立空模組或空介面。

**跨專案邊界**：Watt 不得 import 任何 Taylor 或 Brunel 的程式碼或資料格式。`proc` 模組若參考 Brunel 的 Job Object 做法，應為獨立實作，不建立編譯期依賴。

---

## 15. 修訂記錄

| 版本 | 日期 | 修改內容 | 作者 |
|---|---|---|---|
| v1.0 | 2026-08-07 | 初版建立 | Maze |
| v1.1 | 2026-08-08 | 依 Codex 唯讀規格審查（SR-001～SR-012）修訂：新增 §1.3 名詞定義；修正 G-1／G-2 與 result.json 產出範圍矛盾；補齊 §0 假設影響範圍並新增 A-7～A-9；§4.2 拆分 Ctrl+C 情境、新增 `check --env` 與 result 寫入失敗之對應處理；F-14 補上與 F-13 stdout 衝突之解法；§7.1 補齊靜態驗證規則彙整並修正 Shell 欄位註解；§7.2 新增 R-7／R-8 與 partial result 欄位規則，修正 `exit_code` 可為 null；§7.3 補齊 `EXIT_INTERNAL_ERROR`／`EXIT_ENVIRONMENT_UNAVAILABLE` 觸發條件；§7.4 補齊 Job Object 綁定時序與關閉前確認；修正 §11.1 章節引用錯誤並新增 11 項測試案例；統一 `environment_unavailable` 用詞 | Claude Code（依 Codex 審查結果） |
| v1.2 | 2026-08-08 | 依 v1.1 複審修訂：統一 partial result 的 step 嘗試語意與 `internal_error` 狀態；補齊 R-4／R-8 對 `output_tail`、`resolved_command` 的已知環境值遮罩規則，並同步修正 AC-7 與測試案例；補齊 R-6 semantic 欄位版本規則、R-9 output tail 大小與 UTF-8 邊界；統一無 active step 時的 cancellation 與 `EXIT_CANCELLED`／`EXIT_INTERNAL_ERROR` 對應，明確化 5 秒清理期限與 Job Object close-on-kill 安全清理；強化 P-2／P-3 的 process tree 清理契約；補齊 `watt check` exit code 語意、`EXIT_STEP_FAILED` 的 cwd 啟動錯誤、default pipeline 缺失與空白欄位驗證；修正 §8 流程與 partial result 限定、F-14 invalid pipeline 的 stdout 行為及 §10.2 遺漏 `EXIT_INVALID_PIPELINE` 的範例分支；新增 internal error 測試與 AC-12，更新架構中的 redaction 職責，並修正 §12 的章節引用 | Codex |
| v1.3 | 2026-08-08 | 依 v1.2 複審修訂：R-7 補齊 `cwd` 等 process 從未啟動之 `failed` step 的 `exit_code` 為 `null`，並明訂該情況不保證與 `failed` 狀態同時具備整數 exit code；Partial Result 欄位規則新增對應條目，並明訂 cancellation 確認失敗時該 step 本身仍標記 `cancelled`；A-7 由「待確認」改為「已確認」以消除與 R-9 FROZEN 數值的矛盾；新增 A-10 明確定義 cancellation 確認期限（5 秒），§7.3 不變式與 §4.2／AC-6 統一引用同一定義來源；R-8 補上遮罩最小長度門檻（8 字元）與短值誤傷風險之免責聲明；P-1 補齊 Job Object 建立時須設定 `JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE`，使 P-3 的 close-on-kill 後備手段實際可用；R-6 補充說明本次草稿修訂不觸發 `schema_version` 遞增之理由 | Claude Code（依複審結果） |
