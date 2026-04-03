// Package chissoku は chissoku 子プロセスの起動と stdout 読み取りを担当する。
package chissoku

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
)

// RunOptions は milestone 1 では最小限のフィールドだけ用意してある。
// 後の milestone で設定ファイルに合わせて足していく。
type RunOptions struct {
	// BinPath は chissoku 実行ファイルのパス。
	BinPath string
	// Device はシリアルデバイス（例: /dev/ttyACM0）。
	Device string
}

// ReadOneLine は chissoku を起動し、stdout から最初の 1 行分の JSON を返す。
// chissoku の --stdout.iterations=1 で 1 回出力後に子プロセスが終了する前提。
func ReadOneLine(ctx context.Context, opt RunOptions) ([]byte, error) {
	if opt.BinPath == "" {
		return nil, fmt.Errorf("chissoku: BinPath is empty (set CHISSOKU_BIN)")
	}
	device := opt.Device
	if device == "" {
		device = "/dev/ttyACM0"
	}
	// README: ./chissoku -q /dev/ttyACM0 — 1 行だけ欲しいので iterations=1。待ち時間を抑えるため interval=1。
	args := []string{
		"-q",
		"--stdout.interval=1",
		"--stdout.iterations=1",
		device,
	}
	cmd := exec.CommandContext(ctx, opt.BinPath, args...)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("chissoku: %w (stderr: %s)", err, bytes.TrimSpace(ee.Stderr))
		}
		return nil, fmt.Errorf("chissoku: %w", err)
	}
	return bytes.TrimSpace(out), nil
}

// StdoutReader は milestone 以降で「継続的に stdout を読む」ために使う予定のプレースホルダ。
type StdoutReader struct {
	r io.Reader
}

// NewStdoutReader は TODO(milestone-2 以降): 長時間動かす chissoku から行を読み続けるときに使う。
func NewStdoutReader(r io.Reader) *StdoutReader {
	return &StdoutReader{r: r}
}
