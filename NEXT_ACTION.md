# Watt — 下一步行動

> 僅保留當前有效前線；明確 closeout 時整體重建，不追加歷史。

## 下一個 Session 目標

`docs\spec.md`（v1.3）已由 Maze 確認轉為 Review 狀態。下一步是開始 Phase 1 實作，從 [Issue #2](https://github.com/bext1998/WattCIAutomationEngine/issues/2)（Go module 骨架、CLI 進入點與 build 設定）著手，這是唯一無阻塞前置的 Child Issue。

## 行動（最多 3 項）

1. 依 [Issue #2](https://github.com/bext1998/WattCIAutomationEngine/issues/2) 建立 `go.mod`（Go 1.24.x）、`cmd/watt` cobra 骨架與 static/zero-CGO build 設定，通過 AC-9/AC-10。
2. #2 完成、合併後，依相依關係推進 [#3](https://github.com/bext1998/WattCIAutomationEngine/issues/3)（pipeline 驗證）與 [#4](https://github.com/bext1998/WattCIAutomationEngine/issues/4)（env/cwd），兩者皆只依賴 #2，可平行推進。
3. 之後才進入核心的 [#6](https://github.com/bext1998/WattCIAutomationEngine/issues/6)（Exec Step 執行核心路徑），依相依關係見各 Issue 的「相依關係」欄位或 GitHub 原生 blocked-by 關係。

## 阻塞與待決策

- 無（spec.md 已審查確認，GitHub Issues 已就緒，可直接開始實作）

## 權威連結

- `docs\spec.md`（v1.3，Review）
- Parent Issue：https://github.com/bext1998/WattCIAutomationEngine/issues/1（追蹤全部 Sub-issue 進度，以 GitHub 為工作狀態權威）
