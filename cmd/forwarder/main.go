// Command forwarder は chissoku の出力を Mackerel に送るフォワーダー（開発中）。
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/nakataki17/co2-mackerel/internal/chissoku"
	"github.com/nakataki17/co2-mackerel/internal/reading"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// TODO(milestone-1): 環境変数や定数から BinPath / Device を決める（config は milestone-5 頃）。
	opt := chissoku.RunOptions{
		BinPath: os.Getenv("CHISSOKU_BIN"),
		Device:  getenvDefault("DEVICE", "/dev/ttyACM0"),
	}

	// TODO(milestone-1): chissoku.ReadOneLine で 1 行取得 → reading.ParseLine → 値を fmt.Println などで表示。
	line, err := chissoku.ReadOneLine(ctx, opt)
	if err != nil {
		return err
	}
	r, err := reading.ParseLine(line)
	if err != nil {
		return err
	}
	fmt.Println(r)

	// --- 以下は未着手 milestone 用のメモ（実装するときにこのコメントを消して進める） ---
	// TODO(milestone-2): time.NewTicker で 1 分ごとに同様の読み取り＋print（または chissoku を常駐させて行ストリームを読む）。
	// TODO(milestone-3): Mackerel API へ POST（internal/mackerel などに切り出し）。
	// TODO(milestone-5): gopkg.in/yaml.v3 で config.yaml を読む。
	_ = time.Second

	return nil
}

func getenvDefault(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}
