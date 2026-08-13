# Watt — 下一步行動

> 僅保留當前有效前線；明確 closeout 時整體重建，不追加歷史。

## 目前狀態

`in-progress`。Issue #6（Exec Step 執行核心路徑）已完成並關閉，PR #27 已合併；Issue #9（Environment Diagnostics 與已知環境值 Redaction）是目前唯一已解除前置阻塞的 P0 工作，尚未開始；Parent Issue #1 仍開啟。

## 下一個 Session 目標

完成 [Issue #9](https://github.com/bext1998/WattCIAutomationEngine/issues/9)：讓 `result.json` 的 `environment` 診斷區塊與 `output_tail`／`resolved_command` 的已知環境值遮罩落地，通過 AC-7 secret 不外洩驗收。

## 行動（最多 3 項）

1. 實作 #9：`internal/env` 產出經 redaction 的 environment diagnostics（os/arch、shell_available、resolved_tools、env_var_names，僅名稱不含值）；`internal/runner` 序列化 `output_tail`／`resolved_command` 前，遮罩 effective environment 中長度 ≥8 字元的已知非空 value。
2. 驗證 AC-7、`TestEnvDiagnostics_NoValuesLeaked`、`TestResult_OutputTail_RedactsKnownEnvValues` 通過，並確認遮罩不延伸至即時終端透傳（R-4／R-8 邊界）。
3. #9 完成後，依優先度排入 [#7](https://github.com/bext1998/WattCIAutomationEngine/issues/7) Shell Step 或 [#8](https://github.com/bext1998/WattCIAutomationEngine/issues/8) Windows Job Object／Cancellation（兩者皆 P1，已解除前置阻塞）。

## 阻塞與待決策

- 阻塞：無；#9 的前置 #5、#6 已關閉。
- 已知測試缺口（PR #26 合併前發現，尚未補測）：`watt check --env` 的「偵測缺少 shell」測試未覆蓋「PATH 有 `powershell.exe`（5.1）但無 `pwsh.exe`（7）」這個 spec §4.2／§8.3 明確禁止 fallback 的情境；建議與 #9 一併補上。
- PR #27（Issue #6）審查（pi + Claude 核實，見 [PR #27 留言](https://github.com/bext1998/WattCIAutomationEngine/pull/27#issuecomment-5285418121)）留下的非阻擋技術債：`orchestrator.WattVersion` 寫死 `"dev"`、`runner.resolved_command` 以空白 join args 對含空白參數失真、Run-mode 失敗訊息未帶 step 名稱、`result.Write` 失敗時吞掉原始 step 錯誤，以及三項測試缺口（無效 UTF-8 截斷邊界、`resolved_command` 特殊字元、`result.Write` 失敗路徑）；不影響已合併之驗收條件，留給後續 issue 或 polish 一併處理。

## 權威連結

- `docs\spec.md`（v1.3，Review）
- [Issue #9](https://github.com/bext1998/WattCIAutomationEngine/issues/9)（目前下一個工作前線）
- [Parent Issue #1](https://github.com/bext1998/WattCIAutomationEngine/issues/1)（追蹤全部 Sub-issue 進度，以 GitHub 為工作狀態權威）
- [PR #27](https://github.com/bext1998/WattCIAutomationEngine/pull/27)（最近合併：Exec Step 執行核心路徑，closes #6）
