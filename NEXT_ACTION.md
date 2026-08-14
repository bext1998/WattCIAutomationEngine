# Watt — 下一步行動

> 僅保留當前有效前線；明確 closeout 時整體重建，不追加歷史。

## 目前狀態

`in-progress`。Issue #7（Shell Step 執行，pwsh／cmd）已完成並關閉，PR #28 已合併（squash，經對抗式審查修正一輪後才合併）；Issue #9（Environment Diagnostics 與已知環境值 Redaction，P0）是目前唯一未開始且無前置阻塞的最高優先前線；Issue #8（Windows Job Object／Cancellation，P1）同樣已解除前置阻塞；Parent Issue #1 仍開啟。

## 下一個 Session 目標

完成 [Issue #9](https://github.com/bext1998/WattCIAutomationEngine/issues/9)：讓 `result.json` 的 `environment` 診斷區塊與 `output_tail`／`resolved_command` 的已知環境值遮罩落地，通過 AC-7 secret 不外洩驗收。

## 行動（最多 3 項）

1. 實作 #9：`internal/env` 產出經 redaction 的 environment diagnostics（os/arch、shell_available、resolved_tools、env_var_names，僅名稱不含值）；`internal/runner` 序列化 `output_tail`／`resolved_command` 前，遮罩 effective environment 中長度 ≥8 字元的已知非空 value。
2. 驗證 AC-7、`TestEnvDiagnostics_NoValuesLeaked`、`TestResult_OutputTail_RedactsKnownEnvValues` 通過，並確認遮罩不延伸至即時終端透傳（R-4／R-8 邊界）。
3. #9 完成後，依優先度排入 [#8](https://github.com/bext1998/WattCIAutomationEngine/issues/8) Windows Job Object／Cancellation（P1，已解除前置阻塞）。

## 阻塞與待決策

- 阻塞：無；#9、#8 的前置 Issue 皆已關閉。
- 待規劃：使用者已提議對過去合併的 PR（#20～#28）補一輪總審查（過去 7 個 PR 全數一次過合併、無 GitHub review comment，也都沒跑過對抗式審查），範圍與優先順序待下個 session 討論定案，尚未建立追蹤 Issue。
- 已知測試缺口（PR #26 合併前發現，尚未補測）：`watt check --env` 的「偵測缺少 shell」測試未覆蓋「PATH 有 `powershell.exe`（5.1）但無 `pwsh.exe`（7）」這個 spec §4.2／§8.3 明確禁止 fallback 的情境；建議與 #9 一併補上。
- PR #27（Issue #6）審查（pi + Claude 核實）留下的非阻擋技術債，PR #28 未處理：`orchestrator.WattVersion` 寫死 `"dev"`、`runner.resolved_command`（exec 模式）以空白 join args 對含空白參數失真、`result.Write` 失敗時吞掉原始 step 錯誤，以及既有測試缺口（無效 UTF-8 截斷邊界、`resolved_command` 特殊字元、`result.Write` 失敗路徑）。
- PR #28（Issue #7）對抗式審查留下的非阻擋觀察：`TestMissingPwsh_NoFallbackTo51` 只防得住 PATH 查找式 fallback，防不住「寫死路徑呼叫 `powershell.exe`」這類 fallback（目前程式碼未踩此坑，純屬測試設計殘餘弱點）；`runner.shellArgs()` 對非 `pwsh`/`cmd` 的值回傳空參數，目前僅靠上游 `pipeline.Validate()` 擋住，`runner` 本身無防禦（目前路徑不可觸發）。

## 權威連結

- `docs\spec.md`（v1.3，Review）
- [Issue #9](https://github.com/bext1998/WattCIAutomationEngine/issues/9)（目前下一個工作前線）
- [Parent Issue #1](https://github.com/bext1998/WattCIAutomationEngine/issues/1)（追蹤全部 Sub-issue 進度，以 GitHub 為工作狀態權威）
- [PR #28](https://github.com/bext1998/WattCIAutomationEngine/pull/28)（最近合併：Shell Step 執行核心路徑，closes #7）
