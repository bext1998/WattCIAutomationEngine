# Watt — 下一步行動

> 僅保留當前有效前線；明確 closeout 時整體重建，不追加歷史。

## 目前狀態

`in-progress`。Issue #31（Env 三層合併與 cwd 解析對抗式審查）已完成並關閉，PR #36 已合併（squash，commit `c07ecf4`）。這次審查抓到一個違反確定性核心承諾的真實 bug：`internal/env` 的 `Merge()` 若同一層 env 定義（pipeline-level 或 step-level）內有大小寫不同但視為同一 key 的重複寫法（例如同時寫 `PATH:`／`Path:`），因 Go map 走訪順序隨機化，合併結果在不同次執行間不確定；已修正為合併前先對同層原始 key 做 lexical sort，確定性穩定。經兩輪驗證（pi 審查修正、獨立 subagent 複審，結論 go）。複審附帶明確要求：已依建議建立 [Issue #37](https://github.com/bext1998/WattCIAutomationEngine/issues/37)（`pipeline.Validate()` 應比照 duplicate step name 慣例擋下同層 case-variant env key，P2，security）——這是獨立於本次修正的產品設計決策，不阻塞 #31 的關閉，但需追蹤完成。[Issue #29](https://github.com/bext1998/WattCIAutomationEngine/issues/29)（對抗式審查回溯總覽）**尚未關閉**——還有 [#32](https://github.com/bext1998/WattCIAutomationEngine/issues/32)（Pipeline 資料模型與靜態驗證對抗式審查，P2）一個 Sub-issue 未開始；Parent Issue #1 仍開啟。

## 下一個 Session 目標

依 GitHub 優先級標籤，[Issue #32](https://github.com/bext1998/WattCIAutomationEngine/issues/32)（Pipeline 資料模型與靜態驗證對抗式審查，P2，對應 #3／PR #22）是 Issue #29 總覽列出的最後一個 Sub-issue，也是目前唯一未開始且無前置阻塞的 P2。完成後可關閉 Issue #29 總覽。[Issue #37](https://github.com/bext1998/WattCIAutomationEngine/issues/37)（P2，security）也已解除阻塞，跟 #32 同優先級，排序留待下個 session 依使用者判斷（#37 範圍明確且小，#32 屬於對抗式審查、模式已熟悉）。

## 行動（最多 3 項）

1. 讀 [Issue #32](https://github.com/bext1998/WattCIAutomationEngine/issues/32)、`internal/pipeline/pipeline.go` 及對應測試，比照 PR #35／#36 的模式（先寫審查計劃提示詞交給 pi，pi 審查＋如有問題就修，再送獨立 subagent 複審，結論 go 才合併）。
2. #32 完成後，Issue #29 總覽三個 Sub-issue 全數完成，可關閉。
3. 視使用者排序決定何時處理 [Issue #37](https://github.com/bext1998/WattCIAutomationEngine/issues/37)（`pipeline.Validate()` 補同層 case-variant env key 檢查），不屬於 #29 範圍但同樣是 security 相關的小型獨立任務。

## 阻塞與待決策

- 阻塞：無；#32 前置 Issue（#3／PR #22）已關閉。
- 已知測試缺口（延續自 PR #26／#27／#28，尚未補測）：`watt check --env` 未覆蓋「PATH 有 `powershell.exe`（5.1）但無 `pwsh.exe`（7）」情境（spec §4.2／§8.3 明確禁止 fallback，**仍未補**）；PR #28（Issue #7）非阻擋觀察（`TestMissingPwsh_NoFallbackTo51` 防不住寫死路徑呼叫 `powershell.exe` 的 fallback；`runner.shellArgs()` 對非 `pwsh`/`cmd` 值回傳空參數，目前僅靠 `pipeline.Validate()` 擋住，`runner` 本身無防禦，**仍未補**）。
- PR #27（Issue #6）技術債：`orchestrator.WattVersion` 寫死 `"dev"`、`resolved_command` 空白 join 對含空白參數失真——兩項皆已於 PR #35 審查確認為低風險 diagnostic 顯示問題，可延後。
- **新增（Issue #37，待排程）**：`pipeline.Validate()` 對 `Env` 欄位完全無檢查，不會擋下同一層大小寫重複的 key（例如同時寫 `PATH:`／`Path:`）；目前 `Merge()` 僅靠穩定 tie-break（ASCII 位元組序）避免非決定性，但使用者寫出這種容易搞混的定義不會得到任何警告。範圍明確、風險低，可獨立排程。
- Issue #8 的已知驗證缺口（延續自前次 closeout，尚未回補）：AC-6 checkbox 關閉當下未勾選——取消流程以 `context` cancellation 模擬 Ctrl+C（語意等價），未實際發送 OS 層級 Ctrl+C 訊號端到端驗證，CI 也未針對取消情境跑專門驗證（僅本機 Windows 手動測過）。
- Issue #9／PR #34 殘餘風險：pwsh 版本探測的 `WaitDelay` 修正保證探測呼叫本身會在期限內返回，但不會終止造成卡住的 wrapper 之 grandchild 行程本身。影響範圍小，暫不視為阻塞。
- 對抗式審查回溯進度：[Issue #29](https://github.com/bext1998/WattCIAutomationEngine/issues/29) 總覽尚未關閉；#30、#31 已完成（PR #35、#36），#32（P2）為最後一個 Sub-issue、目前前線。
- 審查過程中額外發現：`docs/spec.md` §4.2 提到的 `--output json` 行為目前 repo 尚未實作，對應既有 [Issue #10](https://github.com/bext1998/WattCIAutomationEngine/issues/10)（P3，非阻塞，純記錄）。

## 權威連結

- `docs\spec.md`（v1.3，Review）
- [Parent Issue #1](https://github.com/bext1998/WattCIAutomationEngine/issues/1)（追蹤全部 Sub-issue 進度，以 GitHub 為工作狀態權威）
- [PR #36](https://github.com/bext1998/WattCIAutomationEngine/pull/36)（最近合併：Env 三層合併與 cwd 解析對抗式審查，closes #31）
- [Issue #29](https://github.com/bext1998/WattCIAutomationEngine/issues/29)（對抗式審查回溯總覽，尚未關閉）
- [Issue #32](https://github.com/bext1998/WattCIAutomationEngine/issues/32)（目前下一個工作前線，P2）
- [Issue #37](https://github.com/bext1998/WattCIAutomationEngine/issues/37)（新建立，P2／security，待排程）
