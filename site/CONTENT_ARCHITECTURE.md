# Watt 文件站內容架構

此記錄實作 Issue #47 的內容權威邊界；它不是公開使用指南，也不是技術契約。

| 位置 | 唯一用途 | 權威範圍 |
| --- | --- | --- |
| `README.md` | 專案入口與短版概覽 | 導向 Releases、文件與技術規格；不作完整手冊 |
| `docs/spec.md` | Watt 的功能、資料模型、Result、exit code 與 process 契約 | 唯一技術契約權威 |
| `site/index.html` | 說明 Watt 的定位、使用者價值與文件入口 | 可白話摘要保證邊界，不得作為契約來源或複製完整 schema／exit code 表 |
| `site/docs/*.html` | 面向使用者的安裝、操作、撰寫與排錯指南 | 非契約；每頁必須標示此身分並連回 `docs/spec.md` |

網站範例只能示範「怎麼使用」，不得讓範例值看起來像 Watt 的固定要求。資料模型、schema、exit code 或任何凍結契約若需要變更，必須先修改 `docs/spec.md` 並走 spec revision；網站只可隨後更新為容易理解的摘要。
