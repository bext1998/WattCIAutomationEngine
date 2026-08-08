# Watt — Repo 結構地圖

> 產出日期：2026-08-09
> 工具：Codex（maze-repo-map）

---

## 技術棧

- **語言**：Go 1.24.x（規格目標；目前尚未建立 Go module）
- **框架**：Cobra CLI（規格目標；目前尚未引入依賴）
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
  docs/spec.md           — Watt Phase 1 v1.3 功能、架構與驗收規格
  .git/                  — Git metadata
```

規格中規劃、但目前倉庫尚不存在的核心路徑：

```
cmd/watt/                — CLI 入口
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
| `NEXT_ACTION.md` | 下一步以 Issue #2 建立 Go module、CLI 骨架與 build 設定 |
| `DECISIONS.md` | 規格狀態與重大設計決策的索引 |
| `AGENTS.md` | 專案範圍、架構方向、執行／authoring 權限與驗證規則 |
| `MAZE_PROJECT.md` | 專案路徑、GitHub repository 與工作流設定 |
| `PROJECT_BRIEF.md` | 專案定位、目標、技術棧與限制摘要 |
| `CLAUDE.md` | Claude Code 工作規則與遠端同步規則 |
| `README.md` | 最小專案簡介 |

---

## 進入點

- **啟動方式**：規格預期以 `watt run [pipeline]`、`watt check [--env]` 執行。
- **主要進入點**：規劃為 `cmd/watt`；目前尚未建立 CLI 實作。

---

## 測試

- **測試目錄**：目前不存在測試目錄或 Go 測試檔。
- **執行測試**：規格預期使用 Go `testing`，涵蓋 unit、integration 與 E2E；目前沒有可執行的測試命令或 `go.mod`。

---

## 設定

- **環境設定**：目前沒有 `.env`、`watt.yaml`、`go.mod` 或其他執行設定檔。
- **關鍵設定項**：規格定義 pipeline 預設檔案為 repo root 的 `watt.yaml`，結果預設寫入 `.watt/result.json`；兩者目前尚未建立。

---

## 目前未知項目

- Issue #2 的 PR／分支內容尚未納入主工作區；需在開始修正前以指定 PR 的 diff 與驗收條件為準。
- `PROJECT_BRIEF.md` 與 `CLAUDE.md` 仍保留「Draft」字樣；`docs/spec.md`、`NEXT_ACTION.md` 與 `MAZE_PROJECT.md` 已標示 v1.3 為 Review，依專案規則以 `docs/spec.md` 為功能權威。
