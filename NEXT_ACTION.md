# Watt — 下一步行動

> 僅保留當前有效前線；明確 closeout 時整體重建，不追加歷史。

## 目前狀態

`completed`（本次 session 範圍）。[Issue #10](https://github.com/bext1998/WattCIAutomationEngine/issues/10)（`--output json` 旗標，§5.2 F-14）已完成並關閉，[PR #39](https://github.com/bext1998/WattCIAutomationEngine/pull/39) 已合併（squash，commit `9678b17`）。實作由 pi（Orca 派工）依 Claude Code 事先確認好的設計完成：`run` 新增 `--output json` 旗標，啟用時 step 即時輸出改導向 stderr、執行結束後把最終 result JSON 寫到真正的 stdout（含 result.json 寫檔失敗仍盡力吐 stdout 的邊界情況），`internal/orchestrator.Outcome` 新增 `Assembled` 欄位判斷是否已組出 result，`internal/result` 抽出共用的 `Marshal()`，未變動 §7.2 FROZEN schema。Issue AC 指定的 `TestOutputJSON_StdoutCarriesOnlyFinalResult` 已新增並通過；Claude Code 獨立重跑過 `go vet ./...`、`go test ./...`（全 7 個套件）確認無回歸；CI `build-and-test` 已通過。Parent [Issue #1](https://github.com/bext1998/WattCIAutomationEngine/issues/1)（Phase 1 MVP 整體追蹤）仍開啟。

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

## 權威連結

- `docs\spec.md`（v1.3，Review）
- [Parent Issue #1](https://github.com/bext1998/WattCIAutomationEngine/issues/1)（追蹤全部 Sub-issue 進度，以 GitHub 為工作狀態權威）
- [PR #39](https://github.com/bext1998/WattCIAutomationEngine/pull/39)（最近合併：`--output json` 旗標，closes #10）
- [Issue #37](https://github.com/bext1998/WattCIAutomationEngine/issues/37)（候選前線，P2／security）
- [Issue #13](https://github.com/bext1998/WattCIAutomationEngine/issues/13)（候選前線，P2／ci，Dogfooding）
