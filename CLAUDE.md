# Watt — Coding Agent 指令

> 本文件供 Claude Code 在每個 session 開始時閱讀。

---

## 專案概述

Watt 是 Windows-first、local-first 的確定性 Pipeline 執行與驗證引擎，讓人類與 coding agent 在本機就能跑完 test / build / package，並產出機器可直接解析的驗證結果。

技術棧：Go 1.24.x + cobra（CLI）+ 檔案系統輸出（`.watt/result.json`），static build、zero CGO。

---

## 工作原則

1. 只實作任務要求的功能，不添加額外功能或重構。
2. 優先編輯現有檔案，只在嚴格必要時建立新檔案。
3. `NEXT_ACTION.md` 是工作狀態權威；只有明確 closeout 才重建它。
4. `docs\spec.md` §7 標記 `[FROZEN]` 的內容（Pipeline 資料模型、Result Schema、Exit Code Contract、Process 管理契約）不得自行增刪改，修改需走 spec revision 並取得使用者確認。

---

## 下一步

閱讀 `NEXT_ACTION.md` 了解這個 session 的目標。

---

## 重要文件

| 文件 | 用途 |
|---|---|
| `docs\spec.md` | 功能規格與驗收標準（v1.3，Draft，待 Maze 正式審查後轉 Review） |
| `NEXT_ACTION.md` | 下一步行動 |
| `DECISIONS.md` | 有效重大決策索引 |
| `PROJECT_BRIEF.md` | 專案一頁式說明 |
| `MAZE_PROJECT.md` | 專案定位與工作流設定 |

---

## 禁止行為

- 不得 force push 到 main / master。
- 不得在使用者未確認前 commit 或 push。
- 不得修改 `docs\spec.md` 的功能範圍或 `[FROZEN]` 契約內容（除非使用者明確要求）。
- spec.md 尚未經 Maze 正式審查確認前，不得視其為定案並據以開始 Phase 1 實作。
