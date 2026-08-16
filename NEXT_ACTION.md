# Watt — 下一步行動

> 僅保留當前有效前線；明確 closeout 時整體重建，不追加歷史。

## 目前狀態

`in-progress`。Issue #8（Windows Job Object 綁定與 Cancellation）已完成並關閉，PR #33 已合併（squash，含一輪審查修正：保留 cleanup 逾時 exit code、補 internal_error 整合測試）；Issue #9（Environment Diagnostics 與已知環境值 Redaction，P0）是目前唯一未開始且無前置阻塞的最高優先前線；Parent Issue #1 仍開啟。

## 下一個 Session 目標

完成 [Issue #9](https://github.com/bext1998/WattCIAutomationEngine/issues/9)：讓 `result.json` 的 `environment` 診斷區塊與 `output_tail`／`resolved_command` 的已知環境值遮罩落地，通過 AC-7 secret 不外洩驗收。

## 行動（最多 3 項）

1. 實作 #9：`internal/env` 產出經 redaction 的 environment diagnostics（os/arch、shell_available、resolved_tools、env_var_names，僅名稱不含值）；`internal/runner` 序列化 `output_tail`／`resolved_command` 前，遮罩 effective environment 中長度 ≥8 字元的已知非空 value。
2. 驗證 AC-7、`TestEnvDiagnostics_NoValuesLeaked`、`TestResult_OutputTail_RedactsKnownEnvValues` 通過，並確認遮罩不延伸至即時終端透傳（R-4／R-8 邊界）。
3. #9 完成後，目前無其他已解除阻塞的 P1；下一步排序（候選任務 vs. #29～#32 對抗式審查回溯）留待下個 session 依使用者優先順序決定。

## 阻塞與待決策

- 阻塞：無；#9 前置 Issue（#5、#6）皆已關閉。
- **新增（Issue #8／PR #33）**：Issue #8 的 AC-6 checkbox 在關閉當下仍是未勾選狀態——取消流程測試以 `context` cancellation 模擬 Ctrl+C（語意等價），未實際發送 OS 層級 Ctrl+C 訊號端到端驗證；GitHub Actions CI 也只驗證了整體 `go test ./...` 通過，沒有取消情境的專門 CI 覆蓋（僅本機 Windows 手動驗證過）。功能本身已合併上線，但這個驗收缺口尚未回補，下次碰 process／cancellation 相關工作時應一併評估是否需要補真實訊號測試。
- 已知測試缺口（延續自 PR #26／#27／#28，尚未補測）：`watt check --env` 未覆蓋「PATH 有 `powershell.exe`（5.1）但無 `pwsh.exe`（7）」情境（spec §4.2／§8.3 明確禁止 fallback）；PR #27（Issue #6）技術債（`orchestrator.WattVersion` 寫死 `"dev"`、exec 模式 `resolved_command` 空白 join 對含空白參數失真、`result.Write` 失敗時吞掉原始 step 錯誤，以及無效 UTF-8 截斷邊界／`resolved_command` 特殊字元／`result.Write` 失敗路徑三項測試缺口）；PR #28（Issue #7）非阻擋觀察（`TestMissingPwsh_NoFallbackTo51` 防不住寫死路徑呼叫 `powershell.exe` 的 fallback；`runner.shellArgs()` 對非 `pwsh`/`cmd` 值回傳空參數，目前僅靠 `pipeline.Validate()` 擋住，`runner` 本身無防禦）。
- 對抗式審查回溯已建立追蹤：[Issue #29](https://github.com/bext1998/WattCIAutomationEngine/issues/29)（總覽）＋ Sub-issue [#30](https://github.com/bext1998/WattCIAutomationEngine/issues/30)（Exec Step／PR #27）、[#31](https://github.com/bext1998/WattCIAutomationEngine/issues/31)（Env 三層合併／PR #23）、[#32](https://github.com/bext1998/WattCIAutomationEngine/issues/32)（Pipeline 資料模型／PR #22）；範圍已定案（不含 #5、#7、CI workflow、Go module 骨架），尚未開始執行，不阻塞 #9。

## 權威連結

- `docs\spec.md`（v1.3，Review）
- [Issue #9](https://github.com/bext1998/WattCIAutomationEngine/issues/9)（目前下一個工作前線）
- [Parent Issue #1](https://github.com/bext1998/WattCIAutomationEngine/issues/1)（追蹤全部 Sub-issue 進度，以 GitHub 為工作狀態權威）
- [PR #33](https://github.com/bext1998/WattCIAutomationEngine/pull/33)（最近合併：Windows Job Object 綁定與 Cancellation，closes #8）
- [Issue #29](https://github.com/bext1998/WattCIAutomationEngine/issues/29)（對抗式審查回溯總覽，非目前前線）
