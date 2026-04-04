// Package chissoku は chissoku を Docker コンテナで起動し、stdout を読み取る。
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

// DefaultDockerImage は chissoku 公式のコンテナイメージ（README の docker run 例に合わせる）。
const DefaultDockerImage = "ghcr.io/northeye/chissoku:latest"

// RunOptions は docker run で chissoku を起動するときの値をまとめる。
type RunOptions struct {
	// DockerImage は ghcr.io/northeye/chissoku 系のイメージ名（タグ含む）。
	DockerImage string
	// DockerExe は docker コマンド（PATH の "docker" でよいときは空）。
	DockerExe string
	// Device はシリアルデバイス（例: /dev/ttyACM0）。--device にも使う。
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

func newCommand(ctx context.Context, opt RunOptions, intervalSec int, iterations int) *exec.Cmd {
	chArgs := buildChissokuArgs(opt, intervalSec, iterations)
	img := strings.TrimSpace(opt.DockerImage)
	if img == "" {
		img = DefaultDockerImage
	}
	dev := opt.devicePath()
	dexe := strings.TrimSpace(opt.DockerExe)
	if dexe == "" {
		dexe = "docker"
	}
	// 公式例: docker run --rm -it --device /dev/ttyACM0:... image [options] /dev/ttyACM0
	// 子プロセスで stdout を読むので -i のみ（-t は不要）
	args := []string{"run", "--rm", "-i", "--device", dev + ":" + dev, img}
	args = append(args, chArgs...)
	return exec.CommandContext(ctx, dexe, args...)
}

// ReadOneLine は chissoku を起動し、stdout から最初の 1 行分の JSON を返す。
func ReadOneLine(ctx context.Context, opt RunOptions) ([]byte, error) {
	cmd := newCommand(ctx, opt, 1, 1)
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
	cmd := newCommand(ctx, opt, opt.intervalSec(), 0)
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
