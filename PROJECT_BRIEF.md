# Watt — 專案說明

> 建立日期：2026-08-08
> 最後更新：2026-08-08

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
- **備用工具**：Codex（本次 spec 審查流程中透過 herdr 互相審查）

---

## 相關文件

- 規格書：docs\spec.md（v1.3，Draft，待 Maze 正式審查後轉 Review）
- 下一步：NEXT_ACTION.md
- 決策紀錄：DECISIONS.md

---

## 重要限制

- 單一 `watt.exe`，static build，zero CGO，不依賴 Node.js / Python 或任何額外 runtime。
- 不提供 Docker / VM / sandbox runner，不提供 filesystem 隔離；Watt 不宣稱任何 sandbox 能力（NG-4）。
- Watt 對任何上層消費者（人類、coding agent、未來的 Taylor）皆無認知，整合方向恆為「消費者 → Watt」。
- spec.md §7 標記 `[FROZEN]` 的介面契約（Pipeline 資料模型、Result Schema、Exit Code Contract、Process 管理契約）未經 spec revision 不得修改。
- spec.md 內仍有兩處工程猜測值（R-8 遮罩最小長度門檻、A-10 cancellation 確認期限）尚待 Maze 確認是否合理，見 NEXT_ACTION.md。
