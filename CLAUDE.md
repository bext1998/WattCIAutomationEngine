# Watt — Claude Code 入口

Watt 是 Windows-first、local-first 的確定性 Pipeline 執行與驗證引擎，讓人類與 coding agent 在本機就能跑完 test / build / package，並產出機器可直接解析的驗證結果。技術棧：Go 1.24.x + cobra（CLI）+ 檔案系統輸出（`.watt/result.json`），static build、zero CGO。

## 權威規則

本專案的工作規則、不可違反事項，以及依任務類型該讀哪份文件的 Routing Table，統一定義於 [`AGENTS.md`](AGENTS.md)（Claude Code 與 Codex 共用同一套規則，不在此重複維護）。開始工作前先讀 `AGENTS.md`，再依其 Routing Table 讀取任務相關文件；不要只讀本檔就動手。

## 下一步

讀 [`NEXT_ACTION.md`](NEXT_ACTION.md) 了解這個 session 的目標。
