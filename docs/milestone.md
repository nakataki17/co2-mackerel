# やること

TODO.mdに書いてある内容はAIが出力したものなので、人間が作る楽しみを失わないため、ちょっとずつ作っていく

## 進め方（Go の練習用）

1. **マイルストーンを 1 つ選ぶ**（下のチェックリストの上からでも、好きな順でも可）。
2. **コード内の `TODO(milestone-N)` を検索**して、コメントの指示どおりに実装する（AI が空の骨組みとコメントを置いてある）。
3. **`go build ./...` と、必要なら実機 / chissoku ありで動作確認**。
4. **人間 or AI にレビュー**してから次の番号へ。

コメントの付け方:

- `TODO(milestone-N): ...` … その番号のマイルストーンで埋める場所（grep しやすいように英語のマーカーに統一）。
- 型やフィールド名・JSON タグなど **仕様どおりに固定した方がよいもの**は先に書いてある。そこは変えず、中身だけ実装する。

主要なファイル:

| パス | 役割 |
|------|------|
| `cmd/forwarder/main.go` | エントリ。各 milestone の配線・ループ。 |
| `internal/chissoku/run.go` | 子プロセス起動・stdout。 |
| `internal/reading/reading.go` | chissoku JSON 1 行のパース。 |
| `internal/mackerel/` | Mackerel サービスメトリクス API への POST。 |

## マイルストーン一覧

- [x] **1** — chissoku を内部で動かし、出力をパースして print
- [x] **2** — 1 分に 1 回値を print（Ticker / 常駐 chissoku など）
- [x] **3** — Mackerel に送信できるようにする
- [ ] **4** — Windows（WSL2）で動作確認
- [ ] **5** — 設定まわりの改善（環境変数の整理・CLI フラグ化など・任意）
- [ ] **6** — Rock 5A（Linux）で動作確認
- [ ] **7** — 簡単にデプロイできるようにする
- [ ] **8** — GitHub に公開する（整備・ライセンス確認など）

# 資料集

## 公式・入門

- [Go の公式ドキュメント（Documentation）](https://go.dev/doc/)
- [A Tour of Go](https://go.dev/tour/) — 文法・並行の入門
- [Effective Go](https://go.dev/doc/effective_go) — 慣用的な書き方
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments) — よくあるスタイル

## このプロジェクトで触る標準パッケージ（pkg.go.dev）

| 用途 | パッケージ |
|------|------------|
| JSON 入出力 | [`encoding/json`](https://pkg.go.dev/encoding/json) |
| HTTP クライアント | [`net/http`](https://pkg.go.dev/net/http) |
| 子プロセス | [`os/exec`](https://pkg.go.dev/os/exec) |
| 終了・タイムアウト | [`context`](https://pkg.go.dev/context) |
| 行読み取り | [`bufio`](https://pkg.go.dev/bufio) |
| 排他・待ち合わせ | [`sync`](https://pkg.go.dev/sync) |
| 定期実行 | [`time`](https://pkg.go.dev/time)（`Ticker`, `Duration`） |
| 環境変数 | [`os`](https://pkg.go.dev/os)（`Getenv`） |
| シグナル | [`os/signal`](https://pkg.go.dev/os/signal) |

## Mackerel API

- [サービスメトリクスを投稿する](https://mackerel.io/ja/api-docs/entry/service-metrics#post-tsdb) — `POST /api/v0/services/{serviceName}/tsdb`
- ベース URL: `https://api.mackerelio.com`

## よく使うコード断片（貼り戻し用）

**JSON タグ（フィールド名を JSON に合わせる）**

```go
type Reading struct {
	CO2   int     `json:"co2"`
	Temp  float64 `json:"temperature"`
}
```

**子プロセス + コンテキスト（親終了時に子も止めやすい）**

```go
cmd := exec.CommandContext(ctx, "/path/to/chissoku", "args...")
```

**一定間隔で処理（送信ループの心臓部）**

```go
t := time.NewTicker(time.Duration(sec) * time.Second)
defer t.Stop()
for {
	select {
	case <-ctx.Done():
		return
	case <-t.C:
		// 送信など
	}
}
```

**HTTP POST（JSON）**

```go
req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
req.Header.Set("Content-Type", "application/json")
req.Header.Set("X-Api-Key", apiKey)
resp, err := httpClient.Do(req)
```

**競合しない共有変数（最新 1 件の保持）**

```go
var mu sync.Mutex
var last *Reading
// Set/Get で mu.Lock / defer mu.Unlock
```

## 並行処理の注意

- 複数 goroutine が同じ変数を触るときは **Mutex** か **channel** で同期する（データ競合は `go test -race` で検出できる）。
- `Ticker` は使い終わったら **`Stop()`** する。