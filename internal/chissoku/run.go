// Package chissoku は chissoku 子プロセスの起動と stdout 読み取りを担当する。
package chissoku

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// RunOptions は chissoku 起動に必要な値をまとめる。
type RunOptions struct {
	// BinPath は chissoku 実行ファイルのパス。
	BinPath string
	// Device はシリアルデバイス（例: /dev/ttyACM0）。
	Device string
	// StdoutIntervalSec は --stdout.interval（秒）。0 または未設定のときは 60（1 分）。
	StdoutIntervalSec int
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

// StreamLines は chissoku を常駐起動し、stdout の 1 行ごとに handle を呼ぶ。
// handle がエラーを返すと子プロセスを終了して StreamLines 全体が終わる。
// ctx がキャンセルされると子プロセスも終了する（exec.CommandContext）。
func StreamLines(ctx context.Context, opt RunOptions, handle func([]byte) error) error {
	if opt.BinPath == "" {
		return fmt.Errorf("chissoku: BinPath is empty (set CHISSOKU_BIN)")
	}
	device := opt.Device
	if device == "" {
		device = "/dev/ttyACM0"
	}
	interval := opt.StdoutIntervalSec
	if interval <= 0 {
		interval = 60
	}
	args := []string{
		"-q",
		fmt.Sprintf("--stdout.interval=%d", interval),
		device,
	}
	cmd := exec.CommandContext(ctx, opt.BinPath, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("chissoku: stdout pipe: %w", err)
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("chissoku: start: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Bytes()
		if err := handle(line); err != nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		_ = cmd.Process.Kill()
		return fmt.Errorf("chissoku: read stdout: %w", err)
	}

	waitErr := cmd.Wait()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if waitErr != nil {
		return fmt.Errorf("chissoku: %w", waitErr)
	}
	return nil
}

// StdoutReader は行ストリームをラップする（テストや拡張用）。
type StdoutReader struct {
	r io.Reader
}

// NewStdoutReader は Reader をラップする。
func NewStdoutReader(r io.Reader) *StdoutReader {
	return &StdoutReader{r: r}
}
