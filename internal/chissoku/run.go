// Package chissoku は chissoku プロセスを起動し、stdout を読み取る。
package chissoku

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// RunOptions は chissoku を起動するときの値をまとめる。
type RunOptions struct {
	// Bin は chissoku 実行ファイルの絶対パス（環境変数 CHISSOKU_BIN）。必須。
	Bin string
	// Device はシリアルデバイス（例: /dev/ttyACM0）。
	Device string
	// StdoutIntervalSec は --stdout.interval（秒）。0 または未設定のときは 60（1 分）。
	StdoutIntervalSec int
}

func (opt RunOptions) devicePath() string {
	if opt.Device != "" {
		return opt.Device
	}
	return "/dev/ttyACM0"
}

func (opt RunOptions) intervalSec() int {
	if opt.StdoutIntervalSec > 0 {
		return opt.StdoutIntervalSec
	}
	return 60
}

// buildChissokuArgs は chissoku に渡す引数（-q, interval, 任意の iterations, デバイスパス）を組み立てる。
// iterations > 0 のときだけ --stdout.iterations を付ける。
func buildChissokuArgs(opt RunOptions, intervalSec int, iterations int) []string {
	dev := opt.devicePath()
	args := []string{"-q", fmt.Sprintf("--stdout.interval=%d", intervalSec)}
	if iterations > 0 {
		args = append(args, fmt.Sprintf("--stdout.iterations=%d", iterations))
	}
	args = append(args, dev)
	return args
}

func newCommand(ctx context.Context, opt RunOptions, intervalSec int, iterations int) (*exec.Cmd, error) {
	bin := strings.TrimSpace(opt.Bin)
	if bin == "" {
		return nil, fmt.Errorf("chissoku: CHISSOKU_BIN is not set (path to the chissoku executable)")
	}
	chArgs := buildChissokuArgs(opt, intervalSec, iterations)
	return exec.CommandContext(ctx, bin, chArgs...), nil
}

func printExec(cmd *exec.Cmd) {
	fmt.Fprintf(os.Stderr, "chissoku: exec %s\n", strings.Join(cmd.Args, " "))
}

// ReadOneLine は chissoku を起動し、stdout から最初の 1 行分の JSON を返す。
func ReadOneLine(ctx context.Context, opt RunOptions) ([]byte, error) {
	cmd, err := newCommand(ctx, opt, 1, 1)
	if err != nil {
		return nil, err
	}
	printExec(cmd)
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
func StreamLines(ctx context.Context, opt RunOptions, handle func([]byte) error) error {
	cmd, err := newCommand(ctx, opt, opt.intervalSec(), 0)
	if err != nil {
		return err
	}
	printExec(cmd)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("chissoku: stdout pipe: %w", err)
	}
	var stderrCopy bytes.Buffer
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderrCopy)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("chissoku: start: %w", err)
	}

	waitAndErr := func(err error) error {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return err
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Bytes()
		if err := handle(line); err != nil {
			return waitAndErr(err)
		}
	}
	if err := scanner.Err(); err != nil {
		return waitAndErr(fmt.Errorf("read stdout: %w", err))
	}

	waitErr := cmd.Wait()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if waitErr != nil {
		return fmt.Errorf("chissoku: %w%s", waitErr, formatStderrHint(stderrCopy.Bytes()))
	}
	return nil
}

// formatStderrHint は chissoku の stderr をエラー文に付ける（空ならヒントだけ）。
func formatStderrHint(stderr []byte) string {
	s := strings.TrimSpace(string(stderr))
	if s != "" {
		return ": " + s
	}
	return " (stderr に詳細なし。CHISSOKU_BIN・DEVICE・シリアル権限（dialout）を確認)"
}

// StdoutReader は行ストリームをラップする（テストや拡張用）。
type StdoutReader struct {
	r io.Reader
}

// NewStdoutReader は Reader をラップする。
func NewStdoutReader(r io.Reader) *StdoutReader {
	return &StdoutReader{r: r}
}
