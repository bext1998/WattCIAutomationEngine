# Execution 模式：用既有 pipeline 驗證修改

## 適用情境
使用者或任務明確要求用既有 pipeline 驗證你剛完成的程式修改，需要獨立於自己判斷的方式驗證是否通過。Repo 根目錄有 `watt.yaml` 本身不是觸發條件；pipeline 定義可執行任意命令，未經明確要求不得執行。

## 流程
1. 完成程式修改後，先跑 `watt check`（無副作用、離線可跑，只做靜態驗證，不啟動任何 step）。若失敗，pipeline 定義本身有問題——回報給使用者，不要自己修 pipeline 定義檔。
2. 跑 `watt run [pipeline] --output json` 並捕獲 stdout。指定 `--output json` 後，若 result 可組裝，stdout 只會有一份本次執行的最終 result JSON（乾淨、可直接 parse），每個 step 的即時輸出改導向 stderr；這份 stdout JSON 才是本次結果，`result.json` 僅是落地副本。
3. 先看 process exit code，這是第一道分流依據，不要先看 JSON 內容判斷：

| exit code | 意義 | 該怎麼處理 |
|---:|---|---|
| 0 | 全部 step 成功 | 驗證通過；若 stdout 有 JSON，保留它作為本次結果 |
| 1 | 有 step 失敗或回傳非零 | 讀取本次執行 stdout 的最終 JSON，查看失敗 step 的 `output_tail.stderr`，去修專案程式碼，絕不修改 pipeline 定義檔；`result.json` 若存在只是落地副本。修完重跑整個 pipeline，不要只重跑失敗的那個 step |
| 2 | pipeline 定義或 CLI 用法錯誤 | 絕不可讀取 `result.json`，因為它可能是舊結果；不要重試同樣的呼叫方式，改讀錯誤訊息（通常會列出可用 pipeline 名稱或行號），這代表設定問題而非程式碼問題 |
| 3 | 缺少必要 command 或 shell（environment_unavailable） | 回報缺少什麼（訊息會明講），並在 stdout 有 JSON 時讀取本次結果；不要用別的 shell 或別的方式硬跑過去——尤其缺 pwsh 時絕對不能 fallback 到 Windows PowerShell 5.1（powershell.exe），語意不保證一致 |
| 4 | 使用者取消（Ctrl+C），process tree 已確認清空 | 非驗證失敗；若 stdout 有 JSON，保留它作為本次結果，回報取消狀態即可 |
| 5 | Watt 自身內部錯誤 | 不得視為驗證失敗；若 stdout 有 JSON，保留它作為本次結果；回報給使用者，這是 Watt 本身的問題，不是專案程式碼的問題 |

4. exit code 判斷完，若 stdout 捕獲到本次執行的最終 JSON，才從該 JSON 的 semantic fields（status、steps[].status、steps[].exit_code）確認細節。`result.json` 僅為落地副本；若寫檔失敗但 result 已組裝，stdout 仍可能有本次 JSON，檔案則可能不存在。若沒有 stdout JSON（例如 invalid pipeline），不得拿既有 `result.json` 猜測本次結果。不得把下列 diagnostic 欄位當判斷依據，它們只供除錯：environment（os/arch/shell_available/resolved_tools/env_var_names）、duration_ms、resolved_command。
5. 幾個容易誤判的欄位語意：
   - environment_unavailable 或 cwd 指向不存在目錄時，該 step 的 exit_code 會是 null——這不是漏欄位，是規格定義的正常值，不要當成解析錯誤。
   - output_tail 只保留每個 stream 尾端最多 8192 bytes，長輸出中間會被截斷；不要假設它是完整輸出。
   - resolved_command／output_tail 裡已知的環境變數值會被遮罩，不代表原始輸出真的是遮罩後的樣子。

## 絕對禁止事項（Execution 模式）
- 修改任何 pipeline 定義檔（watt.yaml 或其等價檔案）——即使只是想讓驗證通過。pipeline 無效就回報設定問題，不得為了通過驗證而改考卷。
- 把非零 exit code 一律當成「程式碼有問題」——exit code 3／4／5 都不是程式碼驗證失敗，分流錯了會誤判。
- parse 終端機的人類可讀輸出取代 `--output json` 捕獲的最終 stdout JSON（例如用字串比對找 "FAIL"）——`--output json` 就是為了避免這樣做。
