# やること

TODO.mdに書いてある内容はAIが出力したものなので、人間が作る楽しみを失わないため、ちょっとずつ作っていく

- [ ] chissokuを内部で動かして、出力をパースしてprintする
- [ ] 1分に1回値をprintするようにする
- [ ] これをmackerelに送信できるようにする
- [ ] Windows(WSL2)環境で動作確認
- [ ] 設定値をいじれるようにする
- [ ] Rock 5a(Linux環境)で動作確認
- [ ] 簡単にデプロイできるようにする
- [ ] GitHubに公開する

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

## 外部ライブラリ（YAML）

- [`gopkg.in/yaml.v3`](https://pkg.go.dev/gopkg.in/yaml.v3) — `config.yaml` の読み込み用（`yaml.Unmarshal`）

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