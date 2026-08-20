# Watt — 有效決策索引

> 只列仍有效、難以逆轉且使用者明確要求同步的決策；細節位於唯一權威 ADR、Issue 或 PR。取代或失效時更新或移除，不追加歷史。

---

## 有效決策

| 摘要 | 狀態 | 唯一權威來源 |
|---|---|---|
| spec.md v1.3 經 Maze 正式審查確認，狀態由 Draft 轉為 Review；§7 `[FROZEN]` 契約自此定案，修改需走 spec revision。v1.4（2026-08-17）依 Issue #41 規格審查修訂 §13 AC-1／§4.1／§10.1 措辭；v1.5（2026-08-19）加強示範文字與契約文字的區隔；v1.6（2026-08-20）依 Issue #50 新增 §5.2 F-25（裸執行 `watt` 品牌橫幅呈現契約）；均經使用者確認，§7 FROZEN 範圍未變動 | 有效 | docs\spec.md 標頭、§7、§15 |
| 其餘重大設計決策（架構選型、shell 預設值、env 策略等）已記錄於 spec.md §8.3、§9，該處即為唯一權威來源 | — | docs\spec.md §8.3、§9 |
| 新建 GitHub Actions CI 禁止使用 Node 20 runtime 的 Action，必須查證後使用 Node 24 相容版本 | 有效 | docs/agent-rules/github-actions.md（源自 Issue #24） |
