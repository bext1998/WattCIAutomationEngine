# MAZE_PROJECT — Watt 定位與工作流設定

> 由 `maze-project-init` 建立。Agent 讀取規格前必須先由此取得實際路徑。
> 文件搬移或設定變更時才更新；不得記錄 token、API key、密碼或私密憑證。

## 專案資訊

- 專案名稱：Watt
- 目標工具：Claude Code（本次 session 內 Codex 亦有參與 spec 審查，視為備用工具）
- 建立日期：2026-08-08

## 文件

- Spec：docs\spec.md（目前版本 v1.3，狀態 Review，Maze 已於 2026-08-08 確認）
- Project Brief：PROJECT_BRIEF.md
- Next Action：NEXT_ACTION.md
- Decisions：DECISIONS.md

## 自適應 Guidance

- Default profile：standard（未特別調整）
- Model overlay：none
- Host capabilities：Claude Code CLI；herdr 多終端協作（可派工至同專案內的 Codex pane）；瀏覽器自動化工具
- Profile escalation evidence：尚無（尚未進入實作）

## GitHub

- Repository：bext1998/WattCIAutomationEngine（https://github.com/bext1998/WattCIAutomationEngine）
- Issue tracking：enabled
- Spec to Issues：enabled
- Priority label convention：repo 目前無 P0–P4 或同義標籤，需新建（見 spec-to-issues Dry Run 的新增標籤建議）
- Category label convention：repo 現有 bug／documentation／enhancement／duplicate／good first issue／help wanted／invalid／question／wontfix；優先比對既有標籤
- Default assignee policy：每次執行 spec-to-issues 時詢問（本人／不指派／逐項／指定帳號）
- Allow label creation：待各次執行時於 Dry Run 取得確認

## 備注

- spec.md 已於 2026-08-08 由 Maze 正式審查確認，狀態轉為 Review；§7 `[FROZEN]` 契約自此視為定案，修改需走 spec revision。
- v1.1～v1.3 的修訂由 Claude Code 與 Codex 透過 herdr 互相審查完成，細節見 spec.md §15 修訂記錄。
- GitHub Issues #1~#18 已依 v1.3 建立（Parent + 12 Child + 5 候選任務），見 https://github.com/bext1998/WattCIAutomationEngine/issues。
