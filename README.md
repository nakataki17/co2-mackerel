# CO2 Mackerel Forwarder

[chissoku](https://github.com/northeye/chissoku) の標準出力（JSON）を読み、[Mackerel](https://mackerel.io/) のサービスメトリクスへ送るフォワーダーです。

## ドキュメント

| ファイル | 内容 |
|----------|------|
| [docs/milestone.md](docs/milestone.md) | マイルストーン・学習用の進め方 |
| [docs/spec.md](docs/spec.md) | 入出力フォーマット・API |

## 前提

- Go（`go.mod` の `go` 行に合わせる）
- **chissoku** の実行ファイル（開発時はリポジトリ直下に置く想定。下記）

## chissoku バイナリ

[リリース](https://github.com/northeye/chissoku/releases)の tar を展開し、**リポジトリ直下**に `chissoku` として置き、`chmod +x chissoku` してください。`.gitignore` 対象のためコミットしません。

## 環境変数

| 変数 | 必須 | 説明 |
|------|------|------|
| `CHISSOKU_BIN` | はい | chissoku のパス（例: `./chissoku`）。リポジトリ直下で動かすときは `./chissoku` または `$PWD/chissoku` |
| `DEVICE` | いいえ | シリアルデバイス。未設定時は `/dev/ttyACM0` |
| `CHISSOKU_INTERVAL_SEC` | いいえ | chissoku の `--stdout.interval`（秒）。未設定時は `60` |

起動時のカレントディレクトリはリポジトリ直下にしてください（`./chissoku` を使う場合）。

## ビルド・テスト・実行

```bash
go build -o bin/forwarder ./cmd/forwarder
go test ./...
```

```bash
export CHISSOKU_BIN="./chissoku"
export CHISSOKU_INTERVAL_SEC=5   # 試すときは短く
go run ./cmd/forwarder
```

`run-local.sh` で同様の環境変数をセットしてから `go run` する例もあります（パスは自分の環境に合わせてください）。

## 設定

本番向けの設定例は [config.example.yaml](config.example.yaml) です（YAML 読み込みは未実装のマイルストーンもあります）。

## ライセンス

MIT License
