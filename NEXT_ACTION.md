# Watt — 下一步行動

> 僅保留當前有效前線；明確 closeout 時整體重建，不追加歷史。

## 目前狀態

`completed`（本次 session 範圍）。官方網站與文件站正式上線（PR #53，Issue #47／#48／#49 全數完成並關閉）：`https://bext1998.github.io/WattCIAutomationEngine/`，`site/CONTENT_ARCHITECTURE.md` 解開 #47 的內容架構決策，GitHub Pages 部署 workflow（`deploy-pages.yml`）已驗證成功。文案「講人話」審查與修正兩輪完成（PR #55，16 處修正，whisk-badger 兩輪複審）。`v0.0.2-snapshot` 已發布（含 ASCII 品牌橫幅 Issue #50），`release.yml` 的 pre-release 旗標 bug 已修正並驗證（PR #52）。新記錄 3 個候選任務（#54 CLI 安裝精靈、#56 Agent Skill 打包、#57 文件站分兩層讀者），皆已跟使用者討論出初步方向但未實作。Parent [Issue #1](https://github.com/bext1998/WattCIAutomationEngine/issues/1)（Phase 1 MVP 整體追蹤）仍開啟。

## 下一個 Session 目標

在「先做無阻塞的 #37」與「挑一個候選任務（#54／#56／#57）深化規劃、開始實作」之間，由使用者選擇下一個工作前線。

## 行動（最多 3 項）

1. 跟使用者確認：先做 #37，還是從 #54／#56／#57 挑一個往下推進。
2. 若選 #37：讀 Issue #37、`internal/pipeline/pipeline.go`（`Validate()`）與 `internal/env`（Issue #31／PR #36 對同層大小寫重複 key 的既有處理方式），補驗證規則與回歸測試。
3. 若選候選任務：三個都還在「已定方向、未定實作規格」階段，需要先把各 Issue 內文的「待討論問題」逐項收斂成可執行範圍，再排入實作。

## 阻塞與待決策

- 待決策：#37 vs 候選任務（#54／#56／#57）由使用者選擇下一個工作前線。
- 已知測試缺口（延續自 PR #26／#27／#28，尚未補測，近期無 PR 觸及）：`watt check --env` 未覆蓋「PATH 有 `powershell.exe`（5.1）但無 `pwsh.exe`（7）」情境；`runner.shellArgs()` 對非 `pwsh`/`cmd` 值目前僅靠 `pipeline.Validate()` 擋住，`runner` 本身無防禦。
- Issue #50／PR #51 已知未驗證項目：ASCII 降級分支（真正 console＋raster／點陣字型）未在真實舊版點陣字型 console 環境測過（Windows 無可靠字型能力偵測 API，設計階段已承認此限制）。
- 主要 checkout（`D:\AgentCoding\WattCIAutomationEngine`）根目錄有一份未追蹤的 `design-qa.md`，來源與用途未確認，closeout 時特意保留未動，未刪除、未 commit。

## 權威連結

- `docs\spec.md`（v1.6，Review）
- [Parent Issue #1](https://github.com/bext1998/WattCIAutomationEngine/issues/1)（追蹤全部 Sub-issue 進度，以 GitHub 為工作狀態權威）
- [Issue #37](https://github.com/bext1998/WattCIAutomationEngine/issues/37)（候選前線，P2／security，無阻塞）
- [Issue #54](https://github.com/bext1998/WattCIAutomationEngine/issues/54)（候選，CLI 安裝精靈：一行安裝指令＋`watt update`子指令，方向已定）／[#56](https://github.com/bext1998/WattCIAutomationEngine/issues/56)（Agent Skill 打包，P2）／[#57](https://github.com/bext1998/WattCIAutomationEngine/issues/57)（文件站分兩層讀者，候選）
- 官方網站：https://bext1998.github.io/WattCIAutomationEngine/
- [PR #53](https://github.com/bext1998/WattCIAutomationEngine/pull/53)（網站上線）、[PR #55](https://github.com/bext1998/WattCIAutomationEngine/pull/55)（文案講人話兩輪修正）——最近合併
