# 測試與驗證規則

> 載入時機：撰寫／修改測試，或判斷一項任務是否已完成驗證。

- 新功能與 bug 修復須依可觀察行為補上最小測試；bug 修復必須包含可重現根因的回歸測試。
- 優先執行受影響 package 的針對性測試，再執行 `go test ./...`。涉及 CLI、process、cancellation、env、result 或跨模組流程時，補跑對應 integration／E2E 測試。
- 涉及 Windows Job Object 或 cancellation 時，必須驗證 child 與 grandchild 全部終止、5 秒確認期限、正常結束無孤兒，以及無法確認時回 `EXIT_INTERNAL_ERROR` 而非 `EXIT_CANCELLED`。
- 涉及 result 或 env 時，必須以 canary 驗證 `.watt/result.json` 全文不洩漏已知環境值，並檢查 `exit_code` nullability、partial steps、UTF-8 合法性及每個 stdout/stderr tail 的 8192-byte 上限。
- 涉及 build 或發布產物時，驗證 `windows/amd64`、`CGO_ENABLED=0`、單一 `watt.exe`，且不依賴 Node.js、Python 或額外 runtime。
- 不得宣稱未執行的檢查已通過。受環境限制無法驗證時，明確列出未驗證項目與風險；GitHub Actions 仍是最終權威驗證。
- 若測試專用 TEMP Go module cache 因唯讀屬性無法清除，在確認該路徑確實為本次測試建立且不包含使用者既有資料後，可解除唯讀屬性再執行清理；不得修改其他路徑的權限或使用高風險系統級刪除操作。

測試策略與案例層級：`docs/spec.md` §11。process／cancellation 契約細節：[`process.md`](process.md)。result／env 契約細節：[`result-env.md`](result-env.md)。
