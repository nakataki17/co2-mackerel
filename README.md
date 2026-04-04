# CO2 Mackerel Forwarder

[chissoku](https://github.com/northeye/chissoku) の標準出力（JSON）を読み、[Mackerel](https://mackerel.io/) のサービスメトリクスへ送るフォワーダーです。

## ドキュメント

| ファイル | 内容 |
|----------|------|
| [docs/deployment.md](docs/deployment.md) | Linux へのデプロイ手順（初心者向け） |
| [docs/milestone.md](docs/milestone.md) | マイルストーン・学習用の進め方 |
| [docs/spec.md](docs/spec.md) | 入出力フォーマット・API・**環境変数一覧** |

## 前提

- Go（`go.mod` の `go` 行に合わせる）
- **chissoku** の実行ファイル（[リリース](https://github.com/northeye/chissoku/releases)から、PC と同じ CPU アーキテクチャ向けを取得）

## 設定

すべて **環境変数** で渡します。名前・既定値・意味は [docs/spec.md](docs/spec.md) の「5. 設定」を参照してください。

API キーは **`MACKEREL_API_KEY`**（固定の変数名）に設定します。

## ビルド・テスト・実行

```bash
go build -o bin/forwarder ./cmd/forwarder
go test ./...
```

```bash
export CHISSOKU_BIN="./chissoku"
export MACKEREL_API_KEY="..."   # Mackerel に送るとき
go run ./cmd/forwarder
```

`run-local.sh` や [docs/deployment.md](docs/deployment.md) の `forwarder.env` 例も参照してください。

## ライセンス

MIT License
