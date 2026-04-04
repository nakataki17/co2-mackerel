# 仕様書

## 1. プロジェクト概要

### 1.1 目的

[chissoku](https://github.com/northeye/chissoku) を使用して取得した CO2 センサデータ（CO2 濃度、温度、湿度）を [Mackerel](https://mackerel.io/) のサービスメトリクスとして送信する。

### 1.2 対象環境

- **ハードウェア**: Rock 5A (RK3588S) または同等の ARM64 SBC
- **OS**: Linux (ARM64)
- **センサ**: UD-CO2S (USB シリアル接続)

## 2. アーキテクチャ

### 2.1 システム構成

```
┌─────────────┐     ┌──────────────────────┐     ┌─────────────┐
│  UD-CO2S    │────▶│ chissoku（Docker 内） │────▶│  forwarder  │
│  (USB)      │     │ docker run … 経由    │     │  (ホスト)   │
└─────────────┘     └──────────────────────┘     └──────┬──────┘
                                                  │
                                                  ▼
                                          ┌─────────────┐
                                          │  Mackerel   │
                                          │     API     │
                                          └─────────────┘
```

### 2.2 コンポーネント

| コンポーネント | 役割 |
|---------------|-----|
| UD-CO2S | CO2 センサ（USB シリアル `/dev/ttyACM0`）|
| chissoku | センサからデータを読み取り JSON を stdout に出す（既定は **Docker コンテナ**で実行） |
| forwarder | `docker run` で chissoku を起動しつつ stdout を読み、Mackerel に送信 |
| Mackerel | メトリクスの可視化・アラート |

### 2.3 データフロー

1. chissoku が `/dev/ttyACM0` からセンサデータを読み取り
2. chissoku が stdout に JSON を出力（デフォルト 60 秒間隔）
3. forwarder が JSON をパースしてメモリに保持
4. forwarder が 60 秒ごとに Mackerel API に POST
5. Mackerel がメトリクスを保存・可視化

## 3. 仕様

### 3.1 入力データ

chissoku の stdout 出力（JSON）:

```json
{
  "co2": 1242,
  "humidity": 31.3,
  "temperature": 29.4,
  "tags": ["Living"],
  "timestamp": "2023-02-01T20:50:51.240+09:00"
}
```

| フィールド | 型 | 説明 |
|-----------|----|----|
| co2 | int | CO2 濃度 (ppm) |
| humidity | float | 湿度 (%) |
| temperature | float | 温度 (°C) |
| tags | []string | タグ（オプション）|
| timestamp | string | ISO 8601 タイムスタンプ |

### 3.2 出力データ

Mackerel Service Metrics API への POST リクエスト:

```json
[
  {
    "name": "co2.living.ppm",
    "time": 1760000000,
    "value": 1242
  },
  {
    "name": "co2.living.temperature_c",
    "time": 1760000000,
    "value": 29.4
  },
  {
    "name": "co2.living.humidity_pct",
    "time": 1760000000,
    "value": 31.3
  }
]
```

### 3.3 メトリクス命名規則

```
co2.{location}.{metric}
```

- `{location}`: 環境変数 `MACKEREL_METRICS_PREFIX` の値に含めて表現する（例: `co2.living` の `living` 部分）
- `{metric}`: `ppm`, `temperature_c`, `humidity_pct`

例: `co2.living.ppm`, `co2.living.temperature_c`, `co2.living.humidity_pct`

### 3.4 API エンドポイント

```
POST https://api.mackerelio.com/api/v0/services/{serviceName}/tsdb
```

ヘッダー:
- `Content-Type`: `application/json`
- `X-Api-Key`: `{MAKEREL_API_KEY}`

## 4. 設計

### 4.1 forwarder の責務

1. **chissoku の起動**: **`docker run`** でコンテナを子プロセスとして起動する
2. **データ読み取り**: stdout から JSON を 1 行ずつ読み取り
3. **データパース**: JSON をパースして構造体に格納
4. **バッファリング**: 最新の測定値をメモリに保持
5. **送信処理**: 60 秒ごとに Mackerel API に POST
6. **エラーハンドリング**: API エラー時のリトライ・ログ出力

### 4.2 プロセス管理

- forwarder 起動時に chissoku を子プロセスとして起動（Docker 経由または直接実行）
- chissoku（コンテナプロセス）が終了した場合は再起動
- systemd から forwarder が終了した場合は chissoku も終了（`exec.CommandContext` 使用）

### 4.3 送信間隔

- **読み取り**: chissoku のデフォルト（60 秒）
- **送信**: 60 秒ごと

Mackerel の推奨する 1 分粒度に合わせ、API 429 を回避する。

### 4.4 エラーハンドリング

| エラー種別 | 対応 |
|-----------|-----|
| chissoku 起動失敗 | 再起動（指数バックオフ）|
| stdout 読み取り失敗 | ログ出力、再起動 |
| JSON パース失敗 | ログ出力、次の行を継続 |
| Mackerel API エラー | リトライ（最大 3 回）|
| ネットワークエラー | 次回の送信タイミングでリトライ |

## 5. 設定

forwarder は **設定ファイルを読みません**。すべて **環境変数** で指定します（systemd の `EnvironmentFile` などにまとめて書ける）。

### 5.0 コマンドライン引数

| 引数 | 説明 |
|------|------|
| `--dry-run` | chissoku（Docker）からセンサ値を読み、**パースして表示するだけ**。Mackerel には **POST しない**（`MACKEREL_API_KEY` があっても送らない）。 |

### 5.1 環境変数

| 変数名 | 説明 | 必須 | 未設定時の挙動 |
|--------|------|------|----------------|
| `CHISSOKU_DOCKER_IMAGE` | chissoku のコンテナイメージ（例: `ghcr.io/northeye/chissoku:latest`） | いいえ | `ghcr.io/northeye/chissoku:latest` |
| `CHISSOKU_DOCKER_EXE` | `docker` 以外のコマンドを使うとき | いいえ | `docker`（`PATH` から検索） |
| `DEVICE` | シリアルデバイス（例: `/dev/ttyACM0`）。`--device` にも使う | いいえ | `/dev/ttyACM0` |
| `CHISSOKU_INTERVAL_SEC` | chissoku の `--stdout.interval`（秒） | いいえ | `60` |
| `MACKEREL_API_KEY` | Mackerel API キー | 送信するなら必須 | 未設定時は POST をスキップ（標準出力のみ） |
| `MACKEREL_SERVICE_NAME` | 投稿先 Mackerel サービス名 | いいえ | `environmental-sensors` |
| `MACKEREL_METRICS_PREFIX` | メトリクス名の接頭辞（例: `co2.living` → `co2.living.ppm` など） | いいえ | `co2.living` |
| `MACKEREL_API_BASE` | Mackerel API のベース URL | いいえ | `https://api.mackerelio.com` |
| `MACKEREL_TIMEOUT_SEC` | HTTP クライアントのタイムアウト（秒） | いいえ | `30` |

メトリクス名は `{MACKEREL_METRICS_PREFIX}.ppm` / `.temperature_c` / `.humidity_pct` の 3 本で固定です。

## 6. 非機能要件

### 6.1 パフォーマンス

- メモリ使用量: < 50 MB
- CPU 使用率: < 5% (アイドル時)
- 起動時間: < 5 秒

### 6.2 可用性

- 自動再起動機能（systemd と forwarder の二重で管理）
- ログ出力（journalctl で確認可能）

### 6.3 セキュリティ

- API キーは環境変数で管理（ファイルに平文保存しない）
- 必要最小限の権限で実行（`dialout` グループのみ）
- 外部接続は Mackerel API のみ

### 6.4 運用

- バージョン管理: symlink 方式でロールバック容易
- 設定変更: 再起動のみで反映
- ログ: systemd journal に統合

## 7. 制約事項

1. **古いデータの扱い**: Mackerel は 24 時間以上前のメトリクスは保存しない
2. **送信レート制限**: 高頻度な POST は 429 エラーの原因になる
3. **単一センサ**: 現バージョンは 1 デバイスのみ対応
4. **ネットワーク依存**: インターネット接続がないと送信失敗（データは破棄）

## 8. 将来の拡張

- [ ] 複数センサ対応
- [ ] ローカルデータ永続化（再送機能）
- [ ] MQTT ブローカー経由の取得
- [ ] メトリクス名のカスタマイズ機能
- [ ] Prometheus exporter 機能
