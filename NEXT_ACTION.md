# Watt — 下一步行動

> 僅保留當前有效前線；明確 closeout 時整體重建，不追加歷史。

## 目前狀態

`completed`（本次 session 範圍）。Issue #32（Pipeline 資料模型與靜態驗證對抗式審查）已完成並關閉，PR #38 已合併（squash，commit `7af2adb`）。這次審查跑了兩輪：pi 第一輪修正 yaml.v3 對 `args`／`env`／step 字串欄位的 scalar coercion 問題（數字、布林值會被靜默轉型成字串，違反 §7.1 FROZEN 資料模型）；獨立 subagent 複審抓到第一輪修正本身引入的新問題——自我參照 YAML merge key（`<<: *self`）會造成無窮遞迴、觸發 Go runtime `fatal error: stack overflow`，是無法被 `recover()` 攔截的行程級當機，轉交 pi 修好（新增 `stringFieldValidator` 做 cycle 偵測）；PR 開出後又跑第三輪獨立 subagent 對 PR 整體做合併前把關審查（額外驗證 pipeline／env 層級的自我參照、間接雙節點循環、深巢狀效能懸崖），結論 go，CI 通過後合併。[Issue #29](https://github.com/bext1998/WattCIAutomationEngine/issues/29)（對抗式審查回溯總覽）三個 Sub-issue（#30／#31／#32）全數完成，已關閉。Parent [Issue #1](https://github.com/bext1998/WattCIAutomationEngine/issues/1)（Phase 1 MVP 整體追蹤）仍開啟。

## 下一個 Session 目標

目前沒有阻塞中的 P0／P1 具體任務（Issue #1 是 Phase 1 整體追蹤，非單一可執行項目）。剩餘最高優先級為 P2，有兩個互不相依的候選，排序留待使用者判斷：

- [Issue #37](https://github.com/bext1998/WattCIAutomationEngine/issues/37)（security，範圍小明確）：`pipeline.Validate()` 應比照 duplicate step name 慣例，擋下同一層大小寫重複的 env key（例如同時寫 `PATH:`／`Path:`）。
- [Issue #13](https://github.com/bext1998/WattCIAutomationEngine/issues/13)（ci）：Dogfooding，讓 Watt 用自己的 `watt.yaml` 跑自身的 test/build/package。

## 行動（最多 3 項）

1. 跟使用者確認先做 #37 還是 #13（或其他方向）。
2. 若選 #37：讀 Issue #37、`internal/pipeline/pipeline.go`（`Validate()`）與 `internal/env`（Issue #31／PR #36 對同層大小寫重複 key 的既有處理方式），補上驗證規則與回歸測試。
3. 若選 #13：讀 Issue #13，規劃 `watt.yaml` 內容涵蓋自身 test/build/package，確認不牴觸 `docs/spec.md` §14 Phase 2 邊界（不擴大 Phase 1 範圍）。

## 阻塞與待決策

- 待決策：#37 vs #13 由使用者選擇下一個工作前線。
- 已知測試缺口（延續自 PR #26／#27／#28，尚未補測）：`watt check --env` 未覆蓋「PATH 有 `powershell.exe`（5.1）但無 `pwsh.exe`（7）」情境（spec §4.2／§8.3 明確禁止 fallback，**仍未補**）；`TestMissingPwsh_NoFallbackTo51` 防不住寫死路徑呼叫 `powershell.exe` 的 fallback，`runner.shellArgs()` 對非 `pwsh`/`cmd` 值回傳空參數目前僅靠 `pipeline.Validate()` 擋住，`runner` 本身無防禦（Issue #7／PR #28 觀察，**仍未補**）。
- PR #27（Issue #6）技術債：`orchestrator.WattVersion` 寫死 `"dev"`、`resolved_command` 空白 join 對含空白參數失真——已確認為低風險 diagnostic 顯示問題，可延後。
- Issue #8 已知驗證缺口：AC-6 取消流程以 `context` cancellation 模擬 Ctrl+C（語意等價），未實際發送 OS 層級 Ctrl+C 訊號端到端驗證，CI 也未針對取消情境跑專門驗證。
- Issue #9／PR #34 殘餘風險：pwsh 版本探測的 `WaitDelay` 修正保證探測呼叫本身會在期限內返回，但不會終止造成卡住的 wrapper 之 grandchild 行程本身，影響範圍小，暫不視為阻塞。
- `docs/spec.md` §4.2 提到的 `--output json` 行為目前 repo 尚未實作，對應既有 [Issue #10](https://github.com/bext1998/WattCIAutomationEngine/issues/10)（P3，非阻塞，純記錄）。

## 權威連結

- `docs\spec.md`（v1.3，Review）
- [Parent Issue #1](https://github.com/bext1998/WattCIAutomationEngine/issues/1)（追蹤全部 Sub-issue 進度，以 GitHub 為工作狀態權威）
- [PR #38](https://github.com/bext1998/WattCIAutomationEngine/pull/38)（最近合併：Pipeline 資料模型與靜態驗證對抗式審查，closes #32）
- [Issue #37](https://github.com/bext1998/WattCIAutomationEngine/issues/37)（候選前線，P2／security）
- [Issue #13](https://github.com/bext1998/WattCIAutomationEngine/issues/13)（候選前線，P2／ci，Dogfooding）
