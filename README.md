# CO2 Mackerel Forwarder

[chissoku](https://github.com/northeye/chissoku) から取得した CO2 センサデータを [Mackerel](https://mackerel.io/) に送信するフォワーダーアプリケーションです。

## 概要

UD-CO2S などの CO2 センサから測定値を取得し、Mackerel のサービスメトリクスとして送信します。ARM64 シングルボードコンピュータ（Rock 5A など）での稼働を想定しています。

## アーキテクチャ

```
UD-CO2S (USB Serial)
    ↓
/dev/ttyACM0
    ↓
chissoku (子プロセス)
    ↓ stdout (JSON)
forwarder
    ↓ HTTPS POST
Mackerel API
```

## 特徴

- **Service Metrics**: ホストメトリクスではなくサービスメトリクスとして送信
- **バッファリング**: 1分ごとにメトリクスをまとめて送信（API 429 回避）
- **自動再起動**: chissoku プロセスが落ちた場合に自動で再起動
- **シンプルなデプロイ**: 実行ファイルのみの配置で動作

## 対応センサ

- UD-CO2S (Seeed Studio)
- chissoku が対応するその他の CO2 センサ

## ディレクトリ構成

本番環境での配置例:

```
/opt/chissoku-forwarder/
  current -> releases/2026-03-26-1200/
  releases/
    2026-03-26-1200/
      chissoku           # chissoku バイナリ
      forwarder          # 本フォワーダー
  config/
    config.yaml

/etc/chissoku-forwarder/
  forwarder.env          # 環境変数（API キーなど）

/etc/systemd/system/
  chissoku-forwarder.service
```

## クイックスタート

詳細なデプロイ手順は [docs/deployment.md](docs/deployment.md) を参照してください。

1. 設定ファイルを作成
2. 実行ファイルを配置
3. systemd サービスを登録・起動

## 設定

設定ファイルの例は [config.yaml.example](config.yaml.example) を参照してください。

## 開発

```bash
# ビルド
go build -o bin/forwarder cmd/forwarder/main.go

# テスト
go test ./...
```

## ライセンス

MIT License
