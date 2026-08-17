# Watt

Watt 是 Windows-first、local-first 的確定性 Pipeline 執行與驗證引擎。它讓人類與 coding agent 能在本機執行 test、build、package，並產出可供程式解析的驗證結果。

Watt 的核心定位是「消費者 → Watt」：Watt 不認知 Taylor、看板、任務系統或呼叫它的人，只負責載入、驗證、依序執行 pipeline，並回傳明確結果。

## 目前狀態

Phase 1 Must Have（`docs/spec.md` §5.1）已全數完成：repository 已合併 Issue #2～#9、#24，包含 Go module、Cobra CLI 入口、Windows/amd64 build 設定、pipeline strict decode／靜態驗證、env/cwd 基礎能力、`watt check --env` 環境探測、`watt run` 的 exec／shell（pwsh、cmd）Step 執行核心路徑、Windows Job Object 綁定與 Ctrl+C cancellation，以及 `result.json` 的 environment 診斷區塊與已知環境值 redaction。`watt --version`、`watt --help`、`watt check`、`watt check --env`、`watt run [pipeline]` 皆可使用；每個 step 皆綁定 Job Object，Ctrl+C 會終止整棵 process tree；`resolved_command`／`output_tail` 序列化前會遮罩已知環境值，`environment` 區塊僅含名稱／OS／arch／shell 版本／工具路徑，不含任何環境變數值。

功能規格位於 [`docs/spec.md`](docs/spec.md)，目前為 v1.3 Review 狀態。Phase 1 Sub-issue 全數完成後，下一步方向待確認（見 [`NEXT_ACTION.md`](NEXT_ACTION.md)）；目前優先級最高的候選項目是 [Issue #29](https://github.com/bext1998/WattCIAutomationEngine/issues/29) 對抗式審查回溯（回顧 Issue #6／#4／#3 已合併程式碼）。

## 技術條件

- Go 1.24.x
- Cobra CLI
- Windows amd64
- static build、`CGO_ENABLED=0`
- 不依賴 Node.js、Python 或其他額外 runtime
- MIT License

Watt 不提供 sandbox、filesystem isolation、secret store、remote runner 或 GitHub Actions 相容層。

## 快速開始

### 安裝

主要方式是到 [GitHub Releases](https://github.com/bext1998/WattCIAutomationEngine/releases) 下載對應版本的 `watt.exe`，並將它放到 PATH 中。

若想執行開發版，可改用 `go install`，但這個方式**需要先自行安裝 Go 1.24 以上版本的工具鏈**（見 [go.dev/dl](https://go.dev/dl/)）——這是 Watt 本身建置階段的依賴，跟建好之後 `watt.exe` 執行時零依賴（不需要 Node.js、Python 或其他 runtime）是兩回事：

```powershell
go install github.com/bext1998/WattCIAutomationEngine/cmd/watt@latest
```

這是輔助安裝方式，不是主要安裝管道；一般使用者請優先用上面的 GitHub Releases 下載，不需要安裝任何額外工具鏈。

### 查看版本與說明

```powershell
go run ./cmd/watt --version
go run ./cmd/watt --help
```

### 執行測試與靜態檢查

```powershell
go test ./...
go vet ./...
```

### 建立 Windows 執行檔

```powershell
.\scripts\build.ps1
.\dist\watt.exe --version
```

Build script 會固定使用 `GOOS=windows`、`GOARCH=amd64` 與 `CGO_ENABLED=0`，產出單一 `dist/watt.exe`。這裡的 `.\dist\watt.exe` 是 repository 內建置產物的相對路徑；使用 Releases 安裝後可直接呼叫 `watt`。

## CLI

目前與 Phase 1 規劃的 CLI 入口如下：

```text
watt --version
watt --help
watt run [pipeline]
watt check
watt check --env
```

目前 `check` 已只做 pipeline 靜態驗證，不啟動任何 step；`check --env` 會額外探測全部 pipeline／step 的 exec 目標與所需 shell（pwsh／cmd）是否可在 PATH 解析，缺失時回 `EXIT_ENVIRONMENT_UNAVAILABLE` 並列出缺項，同樣不啟動任何 process、不寫 result.json。`run` 已依 pipeline 定義循序執行 `exec` 與 `run`（shell：pwsh／cmd）兩種型別 step（fail-fast、result.json 產出），每個 step 皆綁定 Windows Job Object；Ctrl+C 會終止整棵 process tree，5 秒內確認清空回 `EXIT_CANCELLED`，確認不了回 `EXIT_INTERNAL_ERROR`（不謊報 cancelled）；`result.json` 額外含 `environment` 診斷區塊，`resolved_command`／`output_tail` 序列化前已遮罩已知環境值。

## Pipeline 目標格式

Phase 1 的 pipeline 定義預設位於 repository root 的 `watt.yaml`：

```yaml
version: 1

pipelines:
  default:
    steps:
      - name: Test
        exec: go
        args: ["test", "./..."]
```

目前已實作 pipeline 資料模型與靜態驗證；驗證接受規格定義的 `exec`、`pwsh` 與 `cmd` 形式。`exec`（Issue #6）與 `pwsh`／`cmd` shell step（Issue #7）的實際執行皆已完成；完整資料模型、驗證規則與執行語意請以 [`docs/spec.md`](docs/spec.md) §7、§8 為準。

## Result 與 exit code 目標契約

成功執行的 `watt run` 預設會產出 `.watt/result.json`。Result 的 semantic fields 用於機器判定，diagnostic fields 僅供除錯；環境變數值不得進入 result artifact。

Phase 1 的 exit code 契約如下：

| Code | 名稱 | 意義 |
|---:|---|---|
| 0 | `EXIT_SUCCESS` | 所有 step 成功 |
| 1 | `EXIT_STEP_FAILED` | step 執行失敗或回傳非零 |
| 2 | `EXIT_INVALID_PIPELINE`／`EXIT_USAGE` | pipeline 或 CLI 輸入／設定錯誤，不應重試 |
| 3 | `EXIT_ENVIRONMENT_UNAVAILABLE` | 缺少必要 command 或 shell |
| 4 | `EXIT_CANCELLED` | 使用者取消，且 process tree 已安全終止 |
| 5 | `EXIT_INTERNAL_ERROR` | Watt 自身錯誤；不得視為驗證失敗 |

Result schema、partial result、redaction 與 process cleanup 契約以 [`docs/spec.md`](docs/spec.md) §7 為準。

## 架構

```text
cmd/watt → internal/orchestrator → internal/pipeline
                                 → internal/runner → internal/proc
                                                   → internal/env
                                 → internal/result
```

- `cmd/watt`：CLI 參數、旗標與 OS exit code
- `internal/orchestrator`：pipeline 選取、循序執行、fail-fast 與結果決策（已實作，含 cancellation 的三個 context 檢查點）
- `internal/pipeline`：YAML 載入、資料模型與靜態驗證（已實作）
- `internal/runner`：單一步驟執行（exec／shell 兩種模式）、輸出擷取與狀態判定（已實作）
- `internal/env`：host → pipeline → step 的環境合併與 cwd 解析、exec／shell PATH 探測、environment diagnostics 產出與已知值 redaction（已實作）
- `internal/proc`：Windows Job Object 綁定、process tree 終止與 cancellation 確認（已實作）
- `internal/result`：Result 組裝、序列化與寫入（已實作）

模組依賴只能由上而下，Watt 不得反向依賴 Taylor 或其他上層消費者。

## 開發規範

- `docs/spec.md` 是功能、架構與驗收標準的權威來源。
- §7 `[FROZEN]` 契約不得在未經 spec revision 的情況下變更。
- Phase 1 維持循序執行與 fail-fast，不預作平行 step、timeout、matrix 或 sandbox 功能。
- Issue #2～#9、#24（Phase 1 Must Have）已合併；下一步方向待確認，見 [`NEXT_ACTION.md`](NEXT_ACTION.md)。
- 修改前請先閱讀 [`NEXT_ACTION.md`](NEXT_ACTION.md) 與相關 spec 章節。

## License

MIT License。詳見 [`LICENSE`](LICENSE)。
