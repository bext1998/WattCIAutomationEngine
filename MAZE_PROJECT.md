# MAZE_PROJECT — Watt 定位與工作流設定

> 由 `maze-project-init` 建立。Agent 讀取規格前必須先由此取得實際路徑。
> 文件搬移或設定變更時才更新；不得記錄 token、API key、密碼或私密憑證。

## 專案資訊

- 專案名稱：Watt
- 目標工具：Claude Code；Codex 可參與規格審查、實作與 closeout 文件同步
- 建立日期：2026-08-08
- 目前狀態：`in-progress`；Issue #2、#3、#4、#5、#6、#7、#8、#9、#24（Phase 1 Must Have）、#29、#30、#31、#32、#10、#13、#41、#44、#46、#47、#48、#49、#50 已完成；#37（P2／security）待排程；新開候選任務 #54（CLI 安裝精靈）、#56（Agent Skill 打包，P2）、#57（文件站分兩層讀者）
- 最後同步：2026-08-21

## 文件

- Spec：docs\spec.md（目前版本 v1.6，狀態 Review；使用者已於 2026-08-20 確認 v1.6 修訂）
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
- Current delivery：Issue #2、#3、#4、#5、#6、#7、#8、#9、#24（Phase 1 Must Have 全數）、#29、#30、#31、#32、#10、#13、#41、#44、#47、#48、#49、#50 已關閉；PR #21、#22、#23、#25、#26、#27、#28、#33、#34、#35、#36、#38、#39、#40、#42、#43、#45、#46、#51、#52、#53、#55 已合併；Parent Issue #1 仍開啟；Issue #37（P2／security，`pipeline.Validate()` 補同層 case-variant env key 檢查）待排程；官方網站與文件站已上線（https://bext1998.github.io/WattCIAutomationEngine/）；新開候選 Issue #54（CLI 安裝精靈）、#56（Agent Skill 打包，P2）、#57（文件站分兩層讀者）
- CI／QA evidence：repository 現有 `.github/workflows/ci.yml`（push／PR／手動觸發，Windows runner 執行 go vet／go test／build／smoke test），已在 PR #25～#28、#33～#36 與 main push 各成功執行一次；仍沒有獨立 QA 報告

## 備注

- spec.md 已於 2026-08-08 由 Maze 正式審查確認，狀態轉為 Review；§7 `[FROZEN]` 契約自此視為定案，修改需走 spec revision。
- v1.1～v1.3 的修訂由 Claude Code 與 Codex 透過 herdr 互相審查完成；v1.4（2026-08-17）由 Claude Code 依 Issue #41 結構化規格審查結果修訂、經使用者確認，細節見 spec.md §15 修訂記錄。
- GitHub Issues #1～#19 已依 v1.3 建立（Parent + 12 Child + 6 候選任務），見 https://github.com/bext1998/WattCIAutomationEngine/issues。
