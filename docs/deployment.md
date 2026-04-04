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

1. PC の **CPU（アーキテクチャ）** を確認する（**chissoku** と **forwarder** の両方のバイナリ向け）  
2. **chissoku** の Linux 用バイナリをミニ PC に置く（公式リリースや自前ビルドなど）  
3. **forwarder** をビルドし、配置する  
4. ディレクトリを作り、**環境変数ファイル** を置く（`CHISSOKU_BIN` など）  
5. **専用ユーザー**を作り、**シリアル**用 `dialout` グループを付与する  
6. **systemd** に登録して自動起動する  
7. ログと Mackerel で動作を確認する  

---

## 3. アーキテクチャの確認（chissoku・forwarder 用）

**chissoku** も **forwarder** も、ミニ PC 上で動く **ネイティブの Linux バイナリ**です。ビルド／コピーするときは **同じ CPU アーキテクチャ** に合わせます。

```bash
uname -m
```

| 出力の例 | 意味（ざっくり） | クロスビルドするときの `GOARCH` の目安 |
|----------|------------------|----------------------------------------|
| `x86_64` | 一般的な 64bit PC | `amd64` |
| `aarch64` | 64bit ARM など | `arm64` |

**chissoku・forwarder のバイナリ**が PC と違うアーキテクチャだと動きません。

[chissoku](https://github.com/northeye/chissoku) のリリースやビルド手順に従い、対象アーキテクチャ用の実行ファイルを用意してください。配置先の例は次節です。

---

## 4. chissoku バイナリの配置

例として **`/opt/chissoku/chissoku`** に置き、実行権を付けます（パスは任意で、`CHISSOKU_BIN` と一致させます）。

```bash
sudo mkdir -p /opt/chissoku
sudo cp /path/to/chissoku /opt/chissoku/chissoku
sudo chmod +x /opt/chissoku/chissoku
```

センサ接続後、手動で 1 行出るか試す例（`forwarder.env` と同じ `DEVICE` を想定）:

```bash
/opt/chissoku/chissoku -q --stdout.interval=5 /dev/ttyACM0
```

シリアル権限が足りない場合は、試すユーザーが `dialout` に入っているか、`sudo` で確認してください。

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

### 6.2 forwarder バイナリを置く

- `forwarder` → `/opt/chissoku-forwarder/current/forwarder`  
- chissoku は **4 章**のとおり別パス（例: `/opt/chissoku/chissoku`）に置き、`CHISSOKU_BIN` で指定します。

```bash
sudo cp forwarder /opt/chissoku-forwarder/current/forwarder
sudo chmod +x /opt/chissoku-forwarder/current/forwarder
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

# chissoku の実行ファイル（必須）
CHISSOKU_BIN=/opt/chissoku/chissoku

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

## 7. 専用ユーザー・シリアル

センサのデバイス（例: `/dev/ttyACM0`）は、通常 **root か `dialout` グループ**だけが読み書きできます。forwarder（および子プロセスの chissoku）を動かすユーザーは **`dialout`** に入れてください。

### 7.1 ユーザーを作る（例: 名前は `chissoku`）

```bash
sudo useradd --system --home /nonexistent --shell /usr/sbin/nologin chissoku
```

### 7.2 グループに入れる（`dialout`）

```bash
sudo usermod -aG dialout chissoku
```

グループを変えたあと、**既にログイン中のセッション**では反映されないことがあります。サービス再起動後に `systemctl status` で確認してください。

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

**センサ値だけ試す**（Mackerel には送らない）: 同じ `forwarder.env` を読み込んだシェルで  
`set -a && source /etc/chissoku-forwarder/forwarder.env && set +a` のあと  
`/opt/chissoku-forwarder/current/forwarder --dry-run` を実行する（[docs/spec.md](spec.md) の `--dry-run`）。

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
| `cannot execute binary file` | **chissoku** または **forwarder** の **アーキテクチャ**が PC と一致しているか |
| `CHISSOKU_BIN is not set` | `forwarder.env` に **`CHISSOKU_BIN=`** があり、パスが実在するか |
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

## 14. SSH から手元の PC でデプロイ（更新）する

いまのように **手元の PC からミニ PC へ SSH** している場合、次のどちらかが扱いやすいです。

### パターン A: ミニ PC 上にリポジトリを clone しておく（`pull`）

ミニ PC に **Go** と **git** を入れておき、例えば `~/co2-mackerel` に clone します。

手元で（リポジトリのルートにいる想定）:

```bash
chmod +x scripts/deploy-remote.sh
./scripts/deploy-remote.sh pull あなたのユーザー名@ミニPCのホスト名
```

リポジトリを `~/co2-mackerel` 以外に置いている場合は第 3 引数に **リモート側の絶対パス**を渡します。

```bash
./scripts/deploy-remote.sh pull user@mini-pc /home/user/proj/co2-mackerel
```

スクリプトは SSH でリモートに入り、`git pull` → `go build` → `/opt/chissoku-forwarder/current/forwarder` へ配置 → `systemctl restart` まで実行します。`sudo` のパスワードを聞かれる場合は、ミニ PC 上のユーザーに **sudo 権限**がある必要があります。

### パターン B: 手元でビルドしてバイナリだけ送る（`copy`）

ミニ PC に **Go を入れない**運用向けです。手元で **Linux 用**バイナリをビルドし、`scp` で送ります。スクリプトが **`ssh 先の CPU（uname -m）`** を見て `GOARCH` を選びます。

```bash
chmod +x scripts/deploy-remote.sh
./scripts/deploy-remote.sh copy user@mini-pc
```

手元の OS が Linux / macOS / WSL なら通常そのまま動きます。手元が Windows の場合は Git Bash や WSL で実行してください。

### スクリプトの既定値を変えたいとき

環境変数で上書きできます（上級者向け）。

| 変数 | 既定 | 意味 |
|------|------|------|
| `FORWARDER_INSTALL` | `/opt/chissoku-forwarder/current/forwarder` | 配置先のパス |
| `SERVICE` | `chissoku-forwarder` | `systemctl restart` するユニット名 |

例:

```bash
FORWARDER_INSTALL=/usr/local/bin/forwarder SERVICE=my-forwarder ./scripts/deploy-remote.sh copy user@mini-pc
```

### SSH の準備（初心者向けメモ）

- 初回だけ `ssh user@ミニPC` で接続テストし、**ホスト鍵の確認**を済ませておくとスクリプトが止まりにくいです。
- **公開鍵認証**（`ssh-copy-id` など）にしておくと、パスワード入力の手間が減ります。
- スクリプトは **`forwarder` のバイナリと `systemctl restart` だけ**です。`chissoku` バイナリの配置や `forwarder.env` の編集はミニ PC 上で別途行ってください。

---

## 関連ファイル

| パス | 内容 |
|------|------|
| [docs/spec.md](spec.md) | JSON / API / 環境変数の一覧 |
| [contrib/chissoku-forwarder.service](../contrib/chissoku-forwarder.service) | systemd のひな型 |
| [scripts/deploy-remote.sh](../scripts/deploy-remote.sh) | SSH 経由でビルド・配置・再起動する補助スクリプト |
