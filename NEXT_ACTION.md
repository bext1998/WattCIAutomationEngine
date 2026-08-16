# Watt — 下一步行動

> 僅保留當前有效前線；明確 closeout 時整體重建，不追加歷史。

## 目前狀態

`in-progress`。Issue #30（Exec Step 執行核心路徑對抗式審查）已完成並關閉，PR #35 已合併（squash，commit `0b131cd`）。這次審查抓到一個真實的嚴重 bug：`internal/runner` 的 `tailBuffer.String()`（R-9 output_tail 截斷邏輯）在截斷點恰好落在多位元組 UTF-8 字元中間時會無限迴圈、卡死 `watt run`；已修正並經兩輪驗證（先由 pi 執行審查與修正、獨立 subagent 複審，結論 go，含獨立數學推導與實際複現）。Phase 1 Must Have（#2～#9、#24）與這次的 Sub-issue #30 皆已完成；[Issue #29](https://github.com/bext1998/WattCIAutomationEngine/issues/29)（對抗式審查回溯總覽）**尚未關閉**——還有 [#31](https://github.com/bext1998/WattCIAutomationEngine/issues/31)（Env 三層合併與 cwd 解析對抗式審查，P1，另標 security）、[#32](https://github.com/bext1998/WattCIAutomationEngine/issues/32)（Pipeline 資料模型與靜態驗證對抗式審查，P2）兩個 Sub-issue 未開始；Parent Issue #1 仍開啟。

## 下一個 Session 目標

依 GitHub 優先級標籤，[Issue #31](https://github.com/bext1998/WattCIAutomationEngine/issues/31)（Env 三層合併與 cwd 解析對抗式審查，P1，另標 security，對應 #4／PR #23）是目前唯一未開始且無前置阻塞的最高優先前線，也是 Issue #29 總覽列出的第二個 Sub-issue。標了 `security` 值得留意——`internal/env` 是這次已完成的 Issue #9（redaction）跟這次剛審過的 Exec Step 都依賴的共用模組，審查範圍會與已驗證過的部分有重疊，開始前應先確認邊界（不重覆審已經在 #9／#30 驗證過的 redaction／exec 路徑，聚焦三層 env 合併本身與 cwd 解析）。

## 行動（最多 3 項）

1. 讀 [Issue #31](https://github.com/bext1998/WattCIAutomationEngine/issues/31)、`internal/env/merge.go`、`internal/env/cwd.go` 及對應測試，比照 PR #35 的模式（先寫審查計劃提示詞交給 pi，pi 審查＋如有問題就修，再送獨立 subagent 複審，結論 go 才合併）。
2. 審查重點依 Issue #31 原文界定（尚待讀取確認細節）；優先檢查 env 合併的大小寫不分優先序（host→pipeline→step）邊界情況、`cwd` 相對路徑解析的邊界／符號連結／絕對路徑情境。
3. #31 完成後接續 [#32](https://github.com/bext1998/WattCIAutomationEngine/issues/32)（P2，Pipeline 資料模型與靜態驗證對抗式審查），三個 Sub-issue 全部完成後可關閉 Issue #29 總覽。

## 阻塞與待決策

- 阻塞：無；#31 前置 Issue（#4／PR #23）已關閉。
- 已知測試缺口（延續自 PR #26／#27／#28，尚未補測，PR #35 已解決其中兩項——見下方更新）：`watt check --env` 未覆蓋「PATH 有 `powershell.exe`（5.1）但無 `pwsh.exe`（7）」情境（spec §4.2／§8.3 明確禁止 fallback，**仍未補**）；PR #28（Issue #7）非阻擋觀察（`TestMissingPwsh_NoFallbackTo51` 防不住寫死路徑呼叫 `powershell.exe` 的 fallback；`runner.shellArgs()` 對非 `pwsh`/`cmd` 值回傳空參數，目前僅靠 `pipeline.Validate()` 擋住，`runner` 本身無防禦，**仍未補**）。
- **PR #35 已解決**：原 PR #27 技術債列表中的「result.Write 失敗路徑無測試」「無效 UTF-8 截斷邊界無測試」兩項缺口，已由 `TestRun_ResultWriteFailureReturnsInternalError`、`TestRunKeepsOutputTailValidAtMultibyteBoundary` 補齊（後者同時是修正無限迴圈 bug 的回歸測試）。剩餘未補：`orchestrator.WattVersion` 寫死 `"dev"`（純 diagnostic，已確認不影響 semantic verdict）、`resolved_command` 空白 join 對含空白參數失真（純 diagnostic 顯示問題，已用 argv helper 測試確認不影響實際傳給子行程的 argv）——兩項皆已確認為低風險，可延後。
- Issue #8 的已知驗證缺口（延續自前次 closeout，尚未回補）：AC-6 checkbox 關閉當下未勾選——取消流程以 `context` cancellation 模擬 Ctrl+C（語意等價），未實際發送 OS 層級 Ctrl+C 訊號端到端驗證，CI 也未針對取消情境跑專門驗證（僅本機 Windows 手動測過）。
- Issue #9／PR #34 殘餘風險：pwsh 版本探測的 `WaitDelay` 修正保證探測呼叫本身會在期限內返回，但不會終止造成卡住的 wrapper 之 grandchild 行程本身。影響範圍小，暫不視為阻塞。
- 對抗式審查回溯進度：[Issue #29](https://github.com/bext1998/WattCIAutomationEngine/issues/29) 總覽尚未關閉；#30 已完成（PR #35），#31（P1）為下一前線，#32（P2）待其後。
- 審查過程中額外發現：`docs/spec.md` §4.2 提到的 `--output json` 行為目前 repo 尚未實作，對應既有 [Issue #10](https://github.com/bext1998/WattCIAutomationEngine/issues/10)（P3，非阻塞，純記錄）。

## 權威連結

- `docs\spec.md`（v1.3，Review）
- [Parent Issue #1](https://github.com/bext1998/WattCIAutomationEngine/issues/1)（追蹤全部 Sub-issue 進度，以 GitHub 為工作狀態權威）
- [PR #35](https://github.com/bext1998/WattCIAutomationEngine/pull/35)（最近合併：Exec Step 執行核心路徑對抗式審查，closes #30）
- [Issue #29](https://github.com/bext1998/WattCIAutomationEngine/issues/29)（對抗式審查回溯總覽，尚未關閉）
- [Issue #31](https://github.com/bext1998/WattCIAutomationEngine/issues/31)（目前下一個工作前線，P1，security）
