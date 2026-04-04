# デプロイ手順（Linux ミニ PC 向け・初心者向け）

この文書は、**自宅の Linux マシン**に chissoku と forwarder を置き、**systemd** で常時起動するまでの流れを説明します。コマンドは `sudo` が必要なものに注記します。設定は **環境変数** のみです（`EnvironmentFile` に一覧を書きます）。

---

## 1. 必要なもの

| もの | 説明 |
|------|------|
| Linux マシン | 例: ミニ PC、シングルボードコンピュータ（**インターネットに出られる**こと） |
| CO2 センサ | UD-CO2S など、[chissoku が対応する機種](https://github.com/northeye/chissoku) |
| USB 接続 | センサを PC に接続し、シリアルデバイスとして認識されること |
| Mackerel アカウント | [mackerel.io](https://mackerel.io/) で組織を作成し、**サービス**と**サービスメトリクス**が使えること |
| Mackerel API キー | 組織の設定などから発行（**他人に見せない**） |

---

## 2. 全体の流れ（ざっくり）

1. PC の **CPU の種類（アーキテクチャ）** を確認する  
2. **chissoku** と **forwarder** の実行ファイルを、その PC 向けに用意する  
3. ディレクトリを作り、バイナリと **環境変数ファイル** を置く  
4. **専用ユーザー**を作り、シリアルポートを読めるようにする  
5. **systemd** に登録して自動起動する  
6. ログと Mackerel で動作を確認する  

---

## 3. アーキテクチャの確認

ミニ PC で次を実行します。

```bash
uname -m
```

| 出力の例 | 意味（ざっくり） | chissoku のリリース名の目安 |
|----------|------------------|-----------------------------|
| `x86_64` | 一般的な 64bit PC | `linux-amd64` |
| `aarch64` | 64bit ARM（Raspberry Pi 4 以降など） | `linux-arm64` |

**PC と違う名前のバイナリ**を置くと動きません。必ず一致させてください。

---

## 4. chissoku の入手と配置

1. [chissoku の Releases](https://github.com/northeye/chissoku/releases) から、**Linux かつ自分の CPU に合う** zip/tar をダウンロードします。  
2. 中の **`chissoku` という名前の実行ファイル**だけ取り出します。  
3. 後で使うので、いったん分かる場所に置きます（次の「ディレクトリ構成」で正式な場所に移します）。

実行権限を付けます。

```bash
chmod +x chissoku
```

---

## 5. forwarder のビルド（ミニ PC 上で）

ミニ PC に **Go** を入れられる場合、リポジトリを置いてビルドするのが簡単です。

```bash
# 例: ホームにリポジトリを clone 済みとする
cd co2-mackerel
go build -o forwarder ./cmd/forwarder
```

ビルドできた `./forwarder` を、後ほど `/opt/chissoku-forwarder/current/` などにコピーします。

**別のパソコンでビルドする場合**は、ミニ PC と **同じ OS・同じアーキテクチャ**向けにクロスビルドします（上級者向け）。迷う場合は **ミニ PC 上でそのまま `go build`** するのが安全です。

---

## 6. ディレクトリとファイルの配置（例）

ここでは **root 向けの例**として `/opt` を使います。別のパスでも構いませんが、**同じような階層**にまとめると管理しやすいです。

### 6.1 ディレクトリを作る

```bash
sudo mkdir -p /opt/chissoku-forwarder/current
sudo mkdir -p /etc/chissoku-forwarder
```

### 6.2 バイナリを置く

- `forwarder` → `/opt/chissoku-forwarder/current/forwarder`  
- `chissoku` → `/opt/chissoku-forwarder/current/chissoku`  

```bash
sudo cp forwarder /opt/chissoku-forwarder/current/forwarder
sudo cp chissoku /opt/chissoku-forwarder/current/chissoku
sudo chmod +x /opt/chissoku-forwarder/current/forwarder
sudo chmod +x /opt/chissoku-forwarder/current/chissoku
```

### 6.3 環境変数ファイル（秘密と設定をまとめる）

`/etc/chissoku-forwarder/forwarder.env` を **root のみが読める**ようにします。

```bash
sudo nano /etc/chissoku-forwarder/forwarder.env
```

次のように **1 行に 1 つ**、`変数名=値` と書きます（`export` は不要です）。

```env
# Mackerel（必須: 送信したい場合）
MACKEREL_API_KEY=ここにあなたのAPIキー

# chissoku の場所（この例では上記パス）
CHISSOKU_BIN=/opt/chissoku-forwarder/current/chissoku

# シリアルデバイス（多くの Linux で /dev/ttyACM0）
DEVICE=/dev/ttyACM0

# chissoku が JSON を出す間隔（秒）。試すときは 10 など短くてもよい
CHISSOKU_INTERVAL_SEC=60

# Mackerel のサービス名（組織に存在する名前に合わせる）
MACKEREL_SERVICE_NAME=environmental-sensors

# メトリクス名のプレフィックス（例: co2.living.ppm などの前半）
MACKEREL_METRICS_PREFIX=co2.living

# 省略可（省略時は Mackerel の公式 URL）
# MACKEREL_API_BASE=https://api.mackerelio.com

# 省略可（省略時は 30 秒）
# MACKEREL_TIMEOUT_SEC=30
```

保存したら権限を絞ります。

```bash
sudo chmod 600 /etc/chissoku-forwarder/forwarder.env
sudo chown root:root /etc/chissoku-forwarder/forwarder.env
```

> **注意**  
> API キーを Git にコミットしたり、共有したりしないでください。

---

## 7. 専用ユーザーとシリアルポート

センサのデバイス（例: `/dev/ttyACM0`）は、通常 **root か `dialout` グループ**だけが読み書きできます。forwarder を **root 以外**で動かすなら、`dialout` に入れます。

### 7.1 ユーザーを作る（例: 名前は `chissoku`）

```bash
sudo useradd --system --home /nonexistent --shell /usr/sbin/nologin chissoku
```

### 7.2 グループに入れる

```bash
sudo usermod -aG dialout chissoku
```

### 7.3 所有権（必要なら）

`/opt/chissoku-forwarder` をこのユーザーが読めるようにします。

```bash
sudo chown -R chissoku:chissoku /opt/chissoku-forwarder
```

`forwarder.env` は **root のまま 600** で問題ありません（systemd が root で読み込み、環境変数としてサービスに渡されます）。

---

## 8. systemd の登録

### 8.1 ユニットファイルを置く

リポジトリの [`contrib/chissoku-forwarder.service`](../contrib/chissoku-forwarder.service) をコピーします。

```bash
sudo cp /path/to/co2-mackerel/contrib/chissoku-forwarder.service /etc/systemd/system/chissoku-forwarder.service
```

中身の **`ExecStart`** が、次のようになっていることを確認してください（引数は不要で、**バイナリのパスだけ**です）。

```ini
ExecStart=/opt/chissoku-forwarder/current/forwarder
```

`User=` と `Group=` は、先ほどのユーザー（例: `chissoku`）と `dialout` に合わせます。

### 8.2 読み込みと起動

```bash
sudo systemctl daemon-reload
sudo systemctl enable chissoku-forwarder
sudo systemctl start chissoku-forwarder
```

### 8.3 状態の確認

```bash
sudo systemctl status chissoku-forwarder
```

`active (running)` になれば起動できています。

---

## 9. ログの見方

```bash
journalctl -u chissoku-forwarder -f
```

- センサ値が **標準出力に相当する行**として出ていれば、読み取りは成功に近いです。  
- `MACKEREL_API_KEY is not set` と出る場合は、**環境変数ファイルのパス**と **`MACKEREL_API_KEY=` の行**を確認してください。

---

## 10. Mackerel 側の確認

1. Mackerel の **サービス**（`MACKEREL_SERVICE_NAME` と同じ名前）があるか確認する。  
2. **サービスメトリクス**に、`MACKEREL_METRICS_PREFIX` に基づく名前（例: `co2.living.ppm`）が流れてくるか確認する。  

名前やサービスが違うとグラフに出ません。まず **名前の完全一致** を疑ってください。

---

## 11. よくあるつまずき

| 現象 | 確認すること |
|------|----------------|
| `Permission denied`（シリアル） | ユーザーが `dialout` か、`/dev/ttyACM0` のパスが正しいか |
| `cannot execute binary file` | chissoku / forwarder の **アーキテクチャ**が PC と一致しているか |
| Mackerel に出ない | `MACKEREL_API_KEY`、サービス名、メトリクス名、ネットワーク（HTTPS が通るか） |
| すぐ落ちる | `journalctl -u chissoku-forwarder -n 50` でエラー全文を確認 |

---

## 12. アップデートのしかた（概要）

1. 新しい `forwarder` をビルドする。  
2. サービスを止める: `sudo systemctl stop chissoku-forwarder`  
3. バイナリを上書きコピーする。  
4. サービスを始める: `sudo systemctl start chissoku-forwarder`  

バージョンごとにディレクトリを分け、`current` を symlink で切り替える方法もあります（上級者向け）。

---

## 13. アンインストールのしかた（概要）

```bash
sudo systemctl disable --now chissoku-forwarder
sudo rm /etc/systemd/system/chissoku-forwarder.service
sudo systemctl daemon-reload
```

あとは `/opt/chissoku-forwarder` と `/etc/chissoku-forwarder/forwarder.env` を削除すればよいです（API キーが残らないよう注意してください）。

---

## 関連ファイル

| パス | 内容 |
|------|------|
| [docs/spec.md](spec.md) | JSON / API / 環境変数の一覧 |
| [contrib/chissoku-forwarder.service](../contrib/chissoku-forwarder.service) | systemd のひな型 |
