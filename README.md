# CO2 Mackerel Forwarder

[chissoku](https://github.com/northeye/chissoku) の標準出力（JSON）を読み、[Mackerel](https://mackerel.io/) のサービスメトリクスへ送るフォワーダーです。

**chissoku は Docker コンテナ**（例: `ghcr.io/northeye/chissoku:latest`）として起動します。ホストに chissoku の単体バイナリは不要です。

## ドキュメント

| ファイル | 内容 |
|----------|------|
| [docs/deployment.md](docs/deployment.md) | Linux へのデプロイ手順（Docker・**SSH からの更新**含む） |
| [scripts/deploy-remote.sh](scripts/deploy-remote.sh) | SSH でミニ PC に forwarder を載せ替える補助スクリプト |
| [docs/milestone.md](docs/milestone.md) | マイルストーン・学習用の進め方 |
| [docs/spec.md](docs/spec.md) | 入出力フォーマット・API・**環境変数一覧** |

## 前提

- **Docker**（`docker` コマンドが使えること。forwarder 実行ユーザーは `docker` グループなどが必要）
- Go（`go.mod` の `go` 行に合わせて forwarder をビルド）
- シリアルデバイスは `--device` でコンテナに渡す（`DEVICE` 環境変数、既定 `/dev/ttyACM0`）

## 設定

すべて **環境変数** で渡します。一覧は [docs/spec.md](docs/spec.md) の「5. 設定」を参照してください。

API キーは **`MACKEREL_API_KEY`**（固定の変数名）に設定します。

## ビルド・テスト・実行

```bash
go build -o bin/forwarder ./cmd/forwarder
go test ./...
```

```bash
export MACKEREL_API_KEY="..."   # Mackerel に送るとき
go run ./cmd/forwarder
```

[docs/deployment.md](docs/deployment.md) の `forwarder.env` 例も参照してください。

## ライセンス

MIT License
