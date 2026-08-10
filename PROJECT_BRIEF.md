# Watt — 專案說明

> 建立日期：2026-08-08
> 最後更新：2026-08-10

---

## 一句話說明

Watt 是 Windows-first、local-first 的確定性 Pipeline 執行與驗證引擎，讓人類與 coding agent 在本機就能跑完 test / build / package，並產出機器可直接解析的驗證結果。

---

## 核心問題

Watt 解決兩個相鄰但不同的問題：

1. **託管 CI 不可用時的執行能力斷層**：GitHub Actions 服務異常、額度耗盡或網路受阻時，repository 與工具鏈其實都在本機，缺的只是一個能照既有定義把 test / build / package 跑完的執行器。
2. **AI 代理產出的可驗證性缺口**：代理宣稱「已完成」不等於通過驗證，需要一個與代理本身無關的、確定性的判定來源，其結果可被機器直接消費。

---

## 技術棧

- **語言**：Go 1.24.x
- **框架 / 主要套件**：cobra（CLI）；static build、zero CGO
- **資料存儲**：無資料庫；結果輸出至檔案系統 `.watt/result.json`
- **目標平台**：Windows（amd64）；MVP 僅支援 windows/amd64，不產出 linux/darwin binary

---

## Coding Agent 工具

- **主要工具**：Claude Code
- **協作工具**：Codex（參與 spec 審查、實作與 closeout 文件同步）；herdr 可協調多個 agent pane

---

## 相關文件

- 規格書：docs\spec.md（v1.3，Review；Maze 已於 2026-08-08 確認）
- 下一步：NEXT_ACTION.md
- 決策紀錄：DECISIONS.md

---

## 重要限制

- 單一 `watt.exe`，static build，zero CGO，不依賴 Node.js / Python 或任何額外 runtime。
- 不提供 Docker / VM / sandbox runner，不提供 filesystem 隔離；Watt 不宣稱任何 sandbox 能力（NG-4）。
- Watt 對任何上層消費者（人類、coding agent、未來的 Taylor）皆無認知，整合方向恆為「消費者 → Watt」。
- spec.md §7 標記 `[FROZEN]` 的介面契約（Pipeline 資料模型、Result Schema、Exit Code Contract、Process 管理契約）未經 spec revision 不得修改。
- spec.md §0 仍保留待確認的假設欄位；實作須以 v1.3 與 §7 `[FROZEN]` 契約為準，不得自行改寫 R-8 或 A-10 的既定語意。

---

## 目前實作基線

- Issue #2 已完成：Go module、Cobra CLI 入口、Windows/amd64、`CGO_ENABLED=0` build 設定。
- Issue #3 已完成：strict YAML decode、pipeline 資料模型與靜態驗證；`watt check` 只載入／驗證 `watt.yaml`，不啟動 step。
- Issue #4 已完成：host → pipeline → step 的 env 合併（key 不分大小寫）與 step `cwd` 相對 repository root 解析。
- Issue #6（Exec Step／`watt run`）尚未實作；Shell Step、Job Object／cancellation、result／redaction、`check --env` 仍是後續工作。
- Issue #24 已完成：新增 `.github/workflows/ci.yml`（push／PR／手動觸發，Windows runner 執行 go vet／go test／build／smoke test），PR #25 已合併；CI 已在 PR 與 main push 各成功執行一次，但仍沒有獨立 QA 報告。`cmd`、`pipeline`、`env` 已有針對性測試檔。
