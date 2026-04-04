# CO2 Mackerel Forwarder

[chissoku](https://github.com/northeye/chissoku) の標準出力（JSON）を読み、[Mackerel](https://mackerel.io/) のサービスメトリクスへ送るフォワーダーです。

**chissoku はホスト上の Linux 用バイナリ**として置き、環境変数 **`CHISSOKU_BIN`** でパスを指定します。

## ドキュメント

| ファイル | 内容 |
|----------|------|
| [docs/deployment.md](docs/deployment.md) | Linux へのデプロイ手順（**SSH からの更新**含む） |
| [scripts/deploy-remote.sh](scripts/deploy-remote.sh) | SSH でミニ PC に forwarder を載せ替える補助スクリプト |
| [docs/milestone.md](docs/milestone.md) | マイルストーン・学習用の進め方 |
| [docs/spec.md](docs/spec.md) | 入出力フォーマット・API・**環境変数一覧** |

## 前提

- **chissoku** の Linux 用実行ファイル（対象 CPU アーキテクチャ向け）を配置し、`CHISSOKU_BIN` にそのパスを設定する
- Go（`go.mod` の `go` 行に合わせて forwarder をビルド）
- シリアルデバイスは chissoku にデバイスパスで渡す（`DEVICE` 環境変数、既定 `/dev/ttyACM0`）。読み取り権限（例: `dialout`）が必要

## 設定

すべて **環境変数** で渡します。一覧は [docs/spec.md](docs/spec.md) の「5. 設定」を参照してください。

API キーは **`MACKEREL_API_KEY`**（固定の変数名）に設定します。

## ビルド・テスト

```bash
go build -o bin/forwarder ./cmd/forwarder
go test ./...
```

## 実行

forwarder は **設定ファイルを読みません**。渡るのは **環境変数だけ**です。

- **systemd で常時起動**するとき: [docs/deployment.md](docs/deployment.md) のとおり `/etc/chissoku-forwarder/forwarder.env` に変数を書き、ユニットの `EnvironmentFile=` で読み込む（手動で `export` は不要）。
- **手元で `go run` するとき**: 同じ `KEY=value` 形式のファイルを **シェルが `source` して**環境に載せる必要があります。ひな型は [forwarder.env.example](forwarder.env.example)（`MACKEREL_API_KEY` などを編集してから）。

```bash
cp forwarder.env.example forwarder.env   # 初回のみ。forwarder.env を編集（CHISSOKU_BIN・MACKEREL_API_KEY 等）
set -a && source ./forwarder.env && set +a
go run ./cmd/forwarder
```

試運転（Mackerel には送らない）:

```bash
set -a && source ./forwarder.env && set +a
go run ./cmd/forwarder --dry-run
```

変数を個別に `export` して動かすこともできますが、運用では **1 ファイルにまとめて source** か **systemd の EnvironmentFile** に揃えるのが扱いやすいです。

## ライセンス

MIT License
