# Process、Shell、Cancellation 規則

> 載入時機：任務涉及 step 啟動方式、Windows Job Object、cancellation、shell step（`run`/`shell`）。
> 完整契約條文（FROZEN，不得自行變更）：`docs/spec.md` §7.4 Process 管理契約、§8.2 確定性邊界。

- `exec` 模式必須直接啟動目標程式，不得經任何 shell 間接執行（spec P-4）。
- Shell step 預設 shell 為 `pwsh`；缺少 PowerShell 7 時必須回報環境不可用（`EXIT_ENVIRONMENT_UNAVAILABLE`），嚴禁 fallback 至 Windows PowerShell 5.1（spec §8.3、A-9）。
- 每個 step 的 process 必須在執行任何使用者程式碼前綁定 Windows Job Object；建立時即設定 `JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE` 或等效 limit，事後才設定無效（spec P-1）。
- cancellation 時必須終止該 Job Object 內整棵 process tree（含 child 與 descendants）（spec P-2）。
- 關閉 Job Object 前必須確認其內已無存活 process；確認期限為 5 秒（A-10），逾時不得靜默視為已取消，需進入 `EXIT_INTERNAL_ERROR` 清理流程，且清理完成前主行程不得退出（spec P-3、§7.3 不變式）。
- Pipeline 執行必須維持循序與 fail-fast；Phase 1 不得加入平行 step、matrix、timeout 或 `continue-on-error`。

測試要求見 [`testing.md`](testing.md) 的 process／cancellation 段落。
