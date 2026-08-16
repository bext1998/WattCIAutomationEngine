# MAZE_PROJECT — Watt 定位與工作流設定

> 由 `maze-project-init` 建立。Agent 讀取規格前必須先由此取得實際路徑。
> 文件搬移或設定變更時才更新；不得記錄 token、API key、密碼或私密憑證。

## 專案資訊

- 專案名稱：Watt
- 目標工具：Claude Code；Codex 可參與規格審查、實作與 closeout 文件同步
- 建立日期：2026-08-08
- 目前狀態：`in-progress`；Issue #2、#3、#4、#5、#6、#7、#8、#9、#24（Phase 1 Must Have）已完成；目前沒有已解除阻塞的 P0，下一步方向待使用者確認（見 NEXT_ACTION.md）
- 最後同步：2026-08-16

## 文件

- Spec：docs\spec.md（目前版本 v1.3，狀態 Review；Maze 已於 2026-08-08 確認）
- Project Brief：PROJECT_BRIEF.md
- Next Action：NEXT_ACTION.md
- Decisions：DECISIONS.md

## 自適應 Guidance

- Default profile：standard（未特別調整）
- Model overlay：none
- Host capabilities：Claude Code CLI；Codex；herdr 多終端協作；瀏覽器自動化工具
- Profile escalation evidence：尚無（目前未需要調整 profile）

## GitHub

- Repository：bext1998/WattCIAutomationEngine（https://github.com/bext1998/WattCIAutomationEngine）
- Issue tracking：enabled
- Spec to Issues：enabled
- Priority label convention：repo 已建立 P0–P4 標籤（先前「無標籤」記錄已過期，於 spec-to-issues 建立 Issue #29～#32 時新建）
- Category label convention：repo 現有 bug／documentation／enhancement／duplicate／good first issue／help wanted／invalid／question／wontfix；優先比對既有標籤
- Default assignee policy：每次執行 spec-to-issues 時詢問（本人／不指派／逐項／指定帳號）
- Allow label creation：待各次執行時於 Dry Run 取得確認
- Current delivery：Issue #2、#3、#4、#5、#6、#7、#8、#9、#24（Phase 1 Must Have 全數）已關閉；PR #21、#22、#23、#25、#26、#27、#28、#33、#34 已合併；Parent Issue #1 仍開啟，目前沒有已解除阻塞的 P0；P1 候選前線為 Issue #29～#31（對抗式審查回溯）
- CI／QA evidence：repository 現有 `.github/workflows/ci.yml`（push／PR／手動觸發，Windows runner 執行 go vet／go test／build／smoke test），已在 PR #25～#28、#33、#34 與 main push 各成功執行一次；仍沒有獨立 QA 報告

## 備注

- spec.md 已於 2026-08-08 由 Maze 正式審查確認，狀態轉為 Review；§7 `[FROZEN]` 契約自此視為定案，修改需走 spec revision。
- v1.1～v1.3 的修訂由 Claude Code 與 Codex 透過 herdr 互相審查完成，細節見 spec.md §15 修訂記錄。
- GitHub Issues #1～#19 已依 v1.3 建立（Parent + 12 Child + 6 候選任務），見 https://github.com/bext1998/WattCIAutomationEngine/issues。
