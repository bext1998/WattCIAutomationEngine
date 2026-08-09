# Watt — 下一步行動

> 僅保留當前有效前線；明確 closeout 時整體重建，不追加歷史。

## 目前狀態

`in-progress`。Issue #3、#4 的前置能力已完成並關閉，Issue #6 是目前唯一已解除前置阻塞的 P0 核心工作；Parent Issue #1 仍開啟。

## 下一個 Session 目標

完成 [Issue #6](https://github.com/bext1998/WattCIAutomationEngine/issues/6) 的 Exec Step 執行核心路徑，讓已完成的 pipeline 靜態驗證與 env/cwd 基礎能力接上第一條可驗證的 `watt run`、partial result 與 exit-code 流程。

## 行動（最多 3 項）

1. 實作 #6：完成 default／具名 pipeline 選取、循序 fail-fast、直接啟動 Exec Step、即時 stdout/stderr 透傳與 result.json 組裝。
2. 依 #6 的驗收條件補齊並驗證 AC-1～AC-3、partial result、cwd／command failure 與 output tail／exit-code 語意。
3. #6 合併並關閉後，依相依關係推進 [#7](https://github.com/bext1998/WattCIAutomationEngine/issues/7) Shell Step 與 [#8](https://github.com/bext1998/WattCIAutomationEngine/issues/8) Windows Job Object／Cancellation。

## 阻塞與待決策

- 阻塞：無；#6 的前置 #3、#4 已關閉。
- 驗證缺口：repository 沒有 GitHub Actions workflow 或執行紀錄，也沒有 QA 報告；需依 `docs\spec.md` §13 逐項補驗。

## 權威連結

- `docs\spec.md`（v1.3，Review）
- [Issue #6](https://github.com/bext1998/WattCIAutomationEngine/issues/6)（目前下一個工作前線）
- [Parent Issue #1](https://github.com/bext1998/WattCIAutomationEngine/issues/1)（追蹤全部 Sub-issue 進度，以 GitHub 為工作狀態權威）
- [PR #23](https://github.com/bext1998/WattCIAutomationEngine/pull/23)（最近合併的 env/cwd 實作）
