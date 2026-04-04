#!/usr/bin/env bash
# 手元の PC から SSH でミニ PC へ forwarder をデプロイする補助スクリプト。
#
# 前提:
#   - ミニ PC に SSH でログインできる（公開鍵認証を推奨）
#   - 配置先・ユニット名は docs/deployment.md の例と同じ（上書きは環境変数で）
#
# 使い方:
#   chmod +x scripts/deploy-remote.sh
#   ./scripts/deploy-remote.sh pull user@mini-pc [リモートのリポジトリディレクトリ]
#   ./scripts/deploy-remote.sh copy user@mini-pc
#
# pull … ミニ PC 上で git pull と go build（ミニ PC に Go・git・clone 済みリポジトリが必要）
# copy … 手元で Linux 向けにビルドして scp（リモートの uname -m に合わせてクロスビルド）
#
set -euo pipefail

FORWARDER_INSTALL="${FORWARDER_INSTALL:-/opt/chissoku-forwarder/current/forwarder}"
SERVICE="${SERVICE:-chissoku-forwarder}"

usage() {
	echo "Usage: $0 pull <user@host> [remote_repo_dir]" >&2
	echo "       $0 copy <user@host>" >&2
	exit 1
}

cmd="${1:-}"
REMOTE="${2:-}"
[[ -n "$cmd" && -n "$REMOTE" ]] || usage

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

remote_pull_build() {
	local path="$1"
	ssh "$REMOTE" "set -euo pipefail
cd $(printf %q "$path")
git pull
go build -o /tmp/forwarder.new ./cmd/forwarder
sudo install -m 755 /tmp/forwarder.new $(printf %q "$FORWARDER_INSTALL")
sudo systemctl restart $(printf %q "$SERVICE")
echo OK: restarted $SERVICE"
}

case "$cmd" in
pull)
	REMOTE_REPO="${3:-}"
	if [[ -z "$REMOTE_REPO" ]]; then
		# リモートのホーム直下の co2-mackerel を想定
		ssh "$REMOTE" "set -euo pipefail
cd ~/co2-mackerel
git pull
go build -o /tmp/forwarder.new ./cmd/forwarder
sudo install -m 755 /tmp/forwarder.new \"$FORWARDER_INSTALL\"
sudo systemctl restart \"$SERVICE\"
echo OK: restarted $SERVICE"
	else
		remote_pull_build "$REMOTE_REPO"
	fi
	;;
copy)
	arch="$(ssh "$REMOTE" uname -m)"
	case "$arch" in
	x86_64) GOARCH=amd64 ;;
	aarch64) GOARCH=arm64 ;;
	armv7l) GOARCH=arm ;;
	*)
		echo "未対応のアーキテクチャ: $arch（scripts/deploy-remote.sh に追記してください）" >&2
		exit 1
		;;
	esac
	echo "リモート CPU: $arch → GOOS=linux GOARCH=$GOARCH でビルドします"
	TMPBIN="$(mktemp)"
	(
		cd "$REPO_ROOT"
		GOOS=linux GOARCH=$GOARCH CGO_ENABLED=0 go build -o "$TMPBIN" ./cmd/forwarder
	)
	scp "$TMPBIN" "$REMOTE:/tmp/forwarder.new"
	rm -f "$TMPBIN"
	ssh "$REMOTE" "sudo install -m 755 /tmp/forwarder.new \"$FORWARDER_INSTALL\" && sudo systemctl restart \"$SERVICE\" && echo OK"
	;;
*)
	usage
	;;
esac
