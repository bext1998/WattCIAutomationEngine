# GitHub Actions Workflow 規則

> 載入時機：新增或修改 `.github/workflows/**`。
> 來源：Issue #24。

- 所有 JavaScript-based Action 必須使用 Node 24 相容版本（`runs.using: node24`）；嚴禁使用 `runs.using: node20` 的 Action，不得因複製舊版 workflow、教學範例或相容性考量而降級。
- 新增或升版第三方 Action 前，必須查證其目前 runtime，例如：`gh api repos/<owner>/<repo>/contents/action.yml?ref=<tag> --jq .content | base64 -d | grep -A2 "^runs:"`；不得憑記憶或慣例假設版本相容。
- 若目標 Action 尚無 Node 24 相容版本，優先順序為：PowerShell script → Go command → Composite Action；不得以 Node 20 版本作為 fallback。
- 不得因 GitHub Actions runtime 需求，而讓 Watt 本體新增 Node.js runtime 依賴（build、test、CLI 均不得引入 Node.js）。
- 目前 `.github/workflows/ci.yml` 使用的第三方 Action 版本（建立當下已查證為 Node 24）：
  - `actions/checkout@v7.0.1`
  - `actions/setup-go@v7.0.0`
  升版時必須重新查證 runtime 再更新此清單。
