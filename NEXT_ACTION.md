# Watt — 下一步行動

> 僅保留當前有效前線；明確 closeout 時整體重建，不追加歷史。

## 目前狀態

`in-progress`。Issue #5（`watt check --env` 環境探測）已完成並關閉，PR #26 已合併；Issue #6 是目前唯一已解除前置阻塞的 P0 核心工作，尚未開始；Parent Issue #1 仍開啟。

## 下一個 Session 目標

完成 [Issue #6](https://github.com/bext1998/WattCIAutomationEngine/issues/6) 的 Exec Step 執行核心路徑，讓已完成的 pipeline 靜態驗證與 env/cwd 基礎能力接上第一條可驗證的 `watt run`、partial result 與 exit-code 流程。

## 行動（最多 3 項）

1. 實作 #6：完成 default／具名 pipeline 選取、循序 fail-fast、直接啟動 Exec Step、即時 stdout/stderr 透傳與 result.json 組裝。
2. 依 #6 的驗收條件補齊並驗證 AC-1～AC-3、partial result、cwd／command failure 與 output tail／exit-code 語意。
3. #6 合併並關閉後，依相依關係推進 [#7](https://github.com/bext1998/WattCIAutomationEngine/issues/7) Shell Step 與 [#8](https://github.com/bext1998/WattCIAutomationEngine/issues/8) Windows Job Object／Cancellation；Issue #5 已完成，其 Blocks 的 Environment Diagnostics／已知環境值 Redaction sub-issue 亦已解除前置阻塞，可視優先度排入後續前線。

## 阻塞與待決策

- 阻塞：無；#6 的前置 #3、#4 已關閉。
- 驗證缺口：repository 現有 `.github/workflows/ci.yml`，已在 PR #25／#26 與 main push 各成功執行一次（go vet／go test／build／smoke test），但仍沒有針對 Issue #6 驗收條件的執行紀錄或 QA 報告；需依 `docs\spec.md` §13 逐項補驗。
- 已知測試缺口（PR #26 合併前審查發現，暫不補測，登記待下次一併處理）：`watt check --env` 現有的「偵測缺少 shell」測試只驗證 PATH 完全清空的情境，沒有驗證「PATH 內有 `powershell.exe`（5.1）但沒有 `pwsh.exe`（7）」這個 spec §4.2／§8.3 明確禁止 fallback 的具體情境；`internal/env.ResolveExecutable` 目前本身沒有 fallback 邏輯，不是現在就壞的 bug，但這條防線沒有測試守住，之後若被誰加了 fallback 不會被抓到。建議在推進 Environment Diagnostics／已知環境值 Redaction sub-issue（或下次碰 `internal/env`／`cmd/watt` 時）一併補上這個測試案例。

## 權威連結

- `docs\spec.md`（v1.3，Review）
- [Issue #6](https://github.com/bext1998/WattCIAutomationEngine/issues/6)（目前下一個工作前線）
- [Parent Issue #1](https://github.com/bext1998/WattCIAutomationEngine/issues/1)（追蹤全部 Sub-issue 進度，以 GitHub 為工作狀態權威）
- [PR #26](https://github.com/bext1998/WattCIAutomationEngine/pull/26)（最近合併：`watt check --env` 環境探測，closes #5）
