# Authoring 模式：建立新的 pipeline 定義

## 適用情境
任務明確要求建立或修改 pipeline 定義，不論 repo 根目錄是否已經有 `watt.yaml`。不要在其他情境下主動建立或修改 pipeline 定義；僅因檔案存在也不代表可以執行它。

## Pipeline 資料模型（自成一份，不假設你能存取 Watt 原始碼或 spec.md）

```yaml
version: 1               # 必填，固定為整數 1

env:                      # 選填：pipeline 級 env override
  KEY: "value"

pipelines:
  <pipeline名稱>:          # 至少一組，watt run 不帶參數時執行 default
    steps:
      - name: <step名稱>   # 必填，同一 pipeline 內不得重複、不得空白
        exec: <執行檔>      # 與 run 恰擇一
        args: ["arg1"]     # 選填，僅 exec 模式有效
        run: |              # 與 exec 恰擇一：shell script 內容
          多行 shell 指令
        shell: pwsh         # 選填，僅 run 模式有效；不填預設 pwsh；只接受 pwsh 或 cmd，不支援 bash
        cwd: <相對路徑>      # 選填，相對 repo root，預設 repo root
        env:                # 選填：step 級 env override（最高優先）
          KEY: "value"
```

**互斥規則**：exec 與 run 必須恰有一個非空；args 只能搭配 exec；shell 只能搭配 run。違反任何一條，靜態驗證會直接失敗（watt check 回傳非零）。未知欄位一律視為錯誤（strict decoding），拼字打錯會直接失敗，不會靜默失效。

## 流程
1. 分析 repo：這個專案的 test／build／package 各自對應什麼指令？（例如 Node 專案可能是 npm test／npm run build；Go 專案是 go test ./...／go build）
2. 依上面的資料模型草擬 watt.yaml。
3. 只能跑 watt check 驗證語法（無副作用、不啟動任何 process）。`watt.yaml` 是可執行任意命令的定義，不是安全設定檔；絕對不能跑 watt run——那會真的執行 pipeline 裡的指令，Authoring 模式下這是越權。
4. watt check 通過只代表語法合法，不代表這份 pipeline 已經可以拿來當正式驗證關卡。交給人類審核之前，不得讓任何自動化流程依賴這份 pipeline 的結果。

## 反例教訓：不要照抄範例文字裡的具體命名當硬性規則
Watt 專案自己的 dogfooding 過程中踩過這個坑：spec 文件裡有個示範情境用 skills.zip 當範例產物檔名，後續有人把這個檔名字面照抄進另一個專案的驗收條件，結果那個專案明明沒有對應內容，卻硬湊出一個名不符實的檔案。教訓：幫任何專案寫 pipeline 時，產物命名、內容應該對應這個專案實際在做的事，不要照抄別的範例或別的專案的具體命名。

## 絕對禁止事項（Authoring 模式）
- 未經任務明確要求就建立或修改 pipeline 定義檔。
- 跑 watt run 驗證自己剛寫的草稿——只能 watt check。
- 把自己寫的、尚未經人類審核的 pipeline 結果當成任何驗收依據回報給使用者。
