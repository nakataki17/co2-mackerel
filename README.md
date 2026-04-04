# CO2 Mackerel Forwarder

[chissoku](https://github.com/northeye/chissoku) の標準出力（JSON）を読み、[Mackerel](https://mackerel.io/) のサービスメトリクスへ送るフォワーダーです。

## ドキュメント

| ファイル | 内容 |
|----------|------|
| [docs/milestone.md](docs/milestone.md) | マイルストーン・学習用の進め方 |
| [docs/spec.md](docs/spec.md) | 入出力フォーマット・API |

## 前提

- Go（`go.mod` の `go` 行に合わせる）
- **chissoku** の実行ファイル（入手・配置の目安は [config.example.yaml](config.example.yaml) の `chissoku` セクションのコメント）

## 設定

[config.example.yaml](config.example.yaml) を `config.yaml` にコピーし、**ファイル内コメント**を見ながら編集してください。YAML にある項目の説明は README では繰り返しません。

API キーは平文で `config.yaml` に書かず、環境変数 **`MACKEREL_API_KEY`** に渡します（名前は固定）。

※ フォワーダーが `config.yaml` をまだ読み込まない場合は、同等の値を環境変数で渡していることがあります。優先順位は実装に従ってください。

## ビルド・テスト・実行

```bash
go build -o bin/forwarder ./cmd/forwarder
go test ./...
```

```bash
cp config.example.yaml config.yaml
# config.yaml と環境変数（API キー等）を用意
go run ./cmd/forwarder
```

`run-local.sh` などで環境を組み立てる例もあります（パスは自分の環境に合わせてください）。

## ライセンス

MIT License
