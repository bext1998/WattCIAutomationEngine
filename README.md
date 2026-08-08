# Watt

Watt 是 Windows-first、local-first 的確定性 Pipeline 執行與驗證引擎。它讓人類與 coding agent 能在本機執行 test、build、package，並產出可供程式解析的驗證結果。

Watt 的核心定位是「消費者 → Watt」：Watt 不認知 Taylor、看板、任務系統或呼叫它的人，只負責載入、驗證、依序執行 pipeline，並回傳明確結果。

## 目前狀態

目前 repository 已建立 Go module、Cobra CLI 入口、Windows/amd64 build 設定與核心 package 骨架。`watt --version`、`watt --help` 可使用；`run` 與 `check` 仍是後續 Phase 1 實作的命令骨架，尚不可作為正式 validation gate。

功能規格位於 [`docs/spec.md`](docs/spec.md)，目前為 v1.3 Review 狀態。下一階段將依序完成 pipeline 靜態驗證、env/cwd 解析與 step 執行核心。

## 技術條件

- Go 1.24.x
- Cobra CLI
- Windows amd64
- static build、`CGO_ENABLED=0`
- 不依賴 Node.js、Python 或其他額外 runtime
- MIT License

Watt 不提供 sandbox、filesystem isolation、secret store、remote runner 或 GitHub Actions 相容層。

## 快速開始

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

Build script 會固定使用 `GOOS=windows`、`GOARCH=amd64` 與 `CGO_ENABLED=0`，產出單一 `dist/watt.exe`。

## CLI

Phase 1 規劃的 CLI 入口如下：

```text
watt --version
watt --help
watt run [pipeline]
watt check [--env]
```

`run` 預期依 pipeline 定義循序執行 test/build/package；`check` 預期只做 pipeline 靜態驗證，不啟動任何 step。兩者的完整行為仍依 Phase 1 實作進度開放。

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

`exec` step 會直接啟動目標程式，不經 shell；shell step 則支援規格定義的 `pwsh` 與 `cmd`。完整資料模型、驗證規則與執行語意請以 [`docs/spec.md`](docs/spec.md) §7、§8 為準。

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
- `internal/orchestrator`：pipeline 選取、循序執行、fail-fast 與結果決策
- `internal/pipeline`：YAML 載入、資料模型與靜態驗證
- `internal/runner`：單一步驟執行、輸出擷取與狀態判定
- `internal/env`：host → pipeline → step 的環境合併與 diagnostics
- `internal/proc`：Windows Job Object、process tree 與 cancellation
- `internal/result`：Result 組裝、序列化與寫入

模組依賴只能由上而下，Watt 不得反向依賴 Taylor 或其他上層消費者。

## 開發規範

- `docs/spec.md` 是功能、架構與驗收標準的權威來源。
- §7 `[FROZEN]` 契約不得在未經 spec revision 的情況下變更。
- Phase 1 維持循序執行與 fail-fast，不預作平行 step、timeout、matrix 或 sandbox 功能。
- 修改前請先閱讀 [`NEXT_ACTION.md`](NEXT_ACTION.md) 與相關 spec 章節。

## License

MIT License。詳見 [`LICENSE`](LICENSE)。
