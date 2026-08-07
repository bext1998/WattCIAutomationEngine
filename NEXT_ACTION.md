# Watt — 下一步行動

> 僅保留當前有效前線；明確 closeout 時整體重建，不追加歷史。

## 下一個 Session 目標

由 Maze 對 `docs\spec.md`（v1.3）做正式人工審查，確認後將文件狀態由 Draft 轉為 Review，作為進入 Phase 1 實作的前置條件。

## 行動（最多 3 項）

1. Maze 通讀 `docs\spec.md` v1.3，重點確認兩處由 AI 互審過程臨場定案、尚無實作經驗佐證的猜測值是否合理：R-8 的遮罩最小長度門檻（8 字元）、A-10 的 cancellation 確認期限（5 秒）。
2. 確認 G-1／G-2 對「artifact 是否與原 CI 等價」「result.json 產出範圍」的收斂措辭是否符合原始產品意圖。
3. 審查通過後，把 spec.md 標頭狀態改為 Review，並視需要在 §15 修訂記錄補上確認紀錄；之後再啟動 Go module 骨架（`cmd/watt`、`internal/{orchestrator,pipeline,runner,result,env,proc}`）。

## 阻塞與待決策

- spec.md 尚未經 Maze 本人審查確認，§7 `[FROZEN]` 契約在此之前不應視為最終定案，不建議在此之前開始實作。

## 權威連結

- `docs\spec.md`（v1.3）
- `docs\spec.md` §15 修訂記錄（v1.0→v1.3 完整修訂歷程）
