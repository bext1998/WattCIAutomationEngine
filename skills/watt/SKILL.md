---
name: watt
description: 用 Watt CLI（watt check／watt run）驗證或建立本機 test/build/package pipeline。當專案根目錄有 watt.yaml，或使用者要求用 Watt 驗證修改、建立 pipeline 定義時使用。
invocation: both
---

# watt

## 目標
安全地用 Watt CLI 驗證專案的 test/build/package pipeline，或在必要時建立 pipeline 定義，不誤判結果、不逾越權限邊界。

## 前置條件
1. 確認 watt 可用：執行 `watt --version`。不可用時，指引使用者到 Watt repo 的 README 安裝小節（GitHub Releases 為主要方式，`go install github.com/bext1998/WattCIAutomationEngine/cmd/watt@latest` 為輔助方式）；若沒有網路能力，也可以說明手動 fallback：用 Go 工具鏈本機 build（`go build -o watt.exe ./cmd/watt`）後把執行檔複製到 PATH 上任一資料夾，這個方式已經實測驗證可行。
2. 判斷任務屬於哪一種模式：
   - repo 已有 `watt.yaml`，任務是驗證某次修改 → **Execution 模式**，讀 `references/execution.md`
   - repo 沒有 `watt.yaml`，任務明確要求建立 pipeline 定義 → **Authoring 模式**，讀 `references/authoring.md`
   - 不確定屬於哪一種、或兩者都可能 → 先確認 `watt.yaml` 是否存在再判斷，不要用猜的

## 邊界
- 兩種模式不得混用：Execution 模式下絕不修改 pipeline 定義檔；Authoring 模式下產生的定義未經人類審核，不得當作正式驗證關卡。
- Watt 的結果判定永遠先看 process exit code，再看 result.json 的 semantic fields；diagnostic fields（environment、duration_ms 等）不得用於判斷成功或失敗。
