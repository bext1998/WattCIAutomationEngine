# Watt — 下一步行動

> 僅保留當前有效前線；明確 closeout 時整體重建，不追加歷史。

## 目前狀態

`in-progress`。Issue #9（Environment Diagnostics 與已知環境值 Redaction）已完成並關閉，PR #34 已合併（squash，commit `74ab28d`）；過程中先跑過一輪對抗式審查（獨立 subagent，結論 revise，抓到 pwsh 版本探測 timeout 對「grandchild 持有 stdout pipe」無效、`RedactKnownValues` 對重疊已知值遮罩不全兩個有實測重現的真問題），修正並重新驗證後才合併。Phase 1 Must Have（§5.1）的 Sub-issue（#2～#9、#24）目前已全數完成並關閉；Parent Issue #1 仍開啟。

**目前沒有已解除阻塞的 P0 或前置阻塞其他工作的 Issue**——這是這個專案第一次進到這個狀態，下一步不是延續既定前線，而是要做優先順序判斷。

## 下一個 Session 目標（待使用者確認，非既定前線）

依 GitHub 優先級標籤（repo 現已建立 P0～P4 標籤，MAZE_PROJECT.md 先前記錄的「無標籤」已過期），目前唯一未開始且無前置阻塞的 P1 是對抗式審查回溯的第一個 Sub-issue：[Issue #30](https://github.com/bext1998/WattCIAutomationEngine/issues/30)（Exec Step 執行核心路徑對抗式審查，對應 #6／PR #27）。但這是工作性質的轉換（從新功能實作轉為回溯既有已合併程式碼），跟先前每個 session 直接接續 Sub-issue 序號不同，建議下個 session 開始前先跟使用者確認要不要照優先級直接開始 #30，還是要先處理其他候選任務（#10～#19，P2～P3／候選任務，皆非阻塞）。

## 行動（最多 3 項，待確認方向後才適用）

1. 若確認往 #30 走：讀 [Issue #30](https://github.com/bext1998/WattCIAutomationEngine/issues/30)、`internal/runner`（exec 模式）、`internal/orchestrator`、`internal/result`，比照這次 PR #34 的模式（先寫方案，跑對抗式審查，抓到真問題再修）。
2. 若使用者改指定其他方向，依指定 Issue 重新確認優先序，不自動套用 P1 排序。
3. 無論方向為何，開始前先確認是否要一併處理下方「阻塞與待決策」列出的已知測試缺口／技術債，避免與新工作範圍衝突。

## 阻塞與待決策

- 阻塞：無。
- **新增（Issue #9／PR #34 殘餘風險，pi 實作時已自陳）**：pwsh 版本探測的 `WaitDelay` 修正保證探測呼叫本身會在期限內返回，但不會終止造成卡住的 wrapper 之 grandchild 行程本身——該情境下 grandchild 可能短暫繼續在背景執行。不屬於 spec §7.4 Job Object process tree 契約範圍（該契約綁定的是 pipeline step 本身，不是這個診斷用途的探測呼叫），影響範圍小，暫不視為阻塞，但下次動到 `internal/env` 探測邏輯時應留意。
- Issue #8 的已知驗證缺口（延續自前次 closeout，尚未回補）：AC-6 checkbox 關閉當下未勾選——取消流程以 `context` cancellation 模擬 Ctrl+C（語意等價），未實際發送 OS 層級 Ctrl+C 訊號端到端驗證，CI 也未針對取消情境跑專門驗證（僅本機 Windows 手動測過）。
- 已知測試缺口（延續自 PR #26／#27／#28，尚未補測）：`watt check --env` 未覆蓋「PATH 有 `powershell.exe`（5.1）但無 `pwsh.exe`（7）」情境（spec §4.2／§8.3 明確禁止 fallback）；PR #27（Issue #6）技術債（`orchestrator.WattVersion` 寫死 `"dev"`、exec 模式 `resolved_command` 空白 join 對含空白參數失真、`result.Write` 失敗時吞掉原始 step 錯誤，以及無效 UTF-8 截斷邊界／`resolved_command` 特殊字元／`result.Write` 失敗路徑三項測試缺口）；PR #28（Issue #7）非阻擋觀察（`TestMissingPwsh_NoFallbackTo51` 防不住寫死路徑呼叫 `powershell.exe` 的 fallback；`runner.shellArgs()` 對非 `pwsh`/`cmd` 值回傳空參數，目前僅靠 `pipeline.Validate()` 擋住，`runner` 本身無防禦）。
- 對抗式審查回溯已建立追蹤：[Issue #29](https://github.com/bext1998/WattCIAutomationEngine/issues/29)（總覽，P1）＋ Sub-issue [#30](https://github.com/bext1998/WattCIAutomationEngine/issues/30)（Exec Step／PR #27，P1）、[#31](https://github.com/bext1998/WattCIAutomationEngine/issues/31)（Env 三層合併／PR #23，P1，另標 security）、[#32](https://github.com/bext1998/WattCIAutomationEngine/issues/32)（Pipeline 資料模型／PR #22，P2）；範圍已定案（不含 #5、#7、CI workflow、Go module 骨架），尚未開始執行。

## 權威連結

- `docs\spec.md`（v1.3，Review）
- [Parent Issue #1](https://github.com/bext1998/WattCIAutomationEngine/issues/1)（追蹤全部 Sub-issue 進度，以 GitHub 為工作狀態權威）
- [PR #34](https://github.com/bext1998/WattCIAutomationEngine/pull/34)（最近合併：Environment Diagnostics 與已知環境值 Redaction，closes #9）
- [Issue #29](https://github.com/bext1998/WattCIAutomationEngine/issues/29)（對抗式審查回溯總覽，P1，候選下一前線）
- [Issue #30](https://github.com/bext1998/WattCIAutomationEngine/issues/30)（候選下一前線，P1，無前置阻塞）
