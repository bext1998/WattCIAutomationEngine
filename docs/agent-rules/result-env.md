# Result、Env、Redaction 規則

> 載入時機：任務涉及 `result.json` 組裝／序列化、env 三層合併、environment diagnostics、已知環境值遮罩。
> 完整 schema 與欄位定義（FROZEN，不得自行變更）：`docs/spec.md` §7.2 Result Schema。

- env 合併順序為 host → pipeline → step，key 不分大小寫，step 層優先權最高。
- `environment` diagnostics 只能輸出 `os`／`arch`／`shell_available`／`resolved_tools`／`env_var_names`（僅變數名稱）；嚴禁輸出任何環境變數的值，即使經過 hash（spec R-4）。
- 序列化 `resolved_command`、`output_tail` 前，必須遮罩 effective environment 中長度 ≥8 字元的已知非空 value；短於門檻的值不逐字比對遮罩（spec R-8）。遮罩不延伸至即時透傳的終端輸出。
- `duration_ms`、`started_at`、`output_tail`、`resolved_command`、`environment`、`watt_version` 皆為 diagnostic fields，不得用於 verdict、cache key、result hash 或等價比較（spec R-1）。
- `exit_code` 為 semantic 欄位；process 未實際啟動或因 cancellation 未取得 OS exit code 時必須為 `null`（spec R-7）。
- pipeline 已成功載入並通過驗證後，success、step failure、`environment_unavailable`、handled cancellation、可組裝的 internal error 都必須盡力產出 partial result；載入／驗證失敗與 `watt check` 不寫 result（spec R-3）。

完整欄位表、partial result 規則與逐條不變式（R-1～R-9）：`docs/spec.md` §7.2。
