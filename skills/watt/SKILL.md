---
name: watt
description: 用 Watt CLI（watt check／watt run）驗證或建立本機 test/build/package pipeline。只有使用者或任務明確要求用 Watt 驗證修改、建立或修改 pipeline 定義時使用；不因專案狀態或 pipeline 定義檔存在而自動觸發。
invocation: both
---

# watt

## 目標
安全地用 Watt CLI 驗證專案的 test/build/package pipeline，或在必要時建立 pipeline 定義，不誤判結果、不逾越權限邊界。

## 觸發與安全
- 只有使用者或任務明確要求用 Watt 驗證、建立或修改 pipeline 定義時才啟用本 Skill；專案存在 `watt.yaml` 不會自動觸發，也不授權執行 `watt run`。
- 將 `watt.yaml` 視為可執行任意命令的 pipeline 定義，而不是安全設定檔。未有明確要求時不得因檔案存在、內容或 repo 狀態自動執行任何 pipeline。

## 前置條件
1. 在已明確要求使用 Watt 後，確認 watt 可用：執行 `watt --version`。不可用時，指引使用者到 Watt repo 的 README 安裝小節（GitHub Releases 為主要方式，`go install github.com/bext1998/WattCIAutomationEngine/cmd/watt@latest` 為輔助方式）。若沒有網路能力，只有在已 checkout Watt 原始碼且目前目錄是該 checkout 的 repo root 時，才可用既有的 `./scripts/build.ps1` 產出 `dist/watt.exe` 再放入 PATH；不可在任意消費者 repo 直接用 Go build 指令取代正式 build script。
2. 判斷任務屬於哪一種模式：
   - 任務明確要求建立或修改 pipeline 定義（不論是否已有 `watt.yaml`）→ **Authoring 模式**，讀 `references/authoring.md`
   - 使用者或任務明確要求用既有 pipeline 驗證程式修改，且未要求修改 pipeline 定義 → **Execution 模式**，讀 `references/execution.md`
   - 沒有明確要求使用 Watt → 不啟用本 Skill，也不執行 `watt check`／`watt run`

## 邊界
- 兩種模式不得混用：Execution 模式下絕不修改 pipeline 定義檔；Authoring 模式下產生的定義未經人類審核，不得當作正式驗證關卡。
- Watt 的結果判定永遠先看 process exit code。使用 `--output json` 時，若 stdout 捕獲到最終 JSON，它才是本次執行的結果；`result.json` 僅是落地副本，可能因寫檔失敗而不存在。只有在確認是本次執行且已取得結果後，才讀取 JSON 的 semantic fields；diagnostic fields（environment、duration_ms 等）不得用於判斷成功或失敗。
