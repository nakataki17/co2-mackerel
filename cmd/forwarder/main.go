// Command forwarder は chissoku の出力を Mackerel に送るフォワーダー（開発中）。
package main

import (
	"context"
	"fmt"
	"os"

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
	opt := chissoku.RunOptions{
		BinPath: os.Getenv("CHISSOKU_BIN"),
		Device:  getenvDefault("DEVICE", "/dev/ttyACM0"),
	}

	line, err := chissoku.ReadOneLine(ctx, opt)
	if err != nil {
		return err
	}
	r, err := reading.ParseLine(line)
	if err != nil {
		return err
	}
	fmt.Printf("CO2=%d ppm, 温度=%.1f°C, 湿度=%.1f%%, 時刻=%s\n",
		r.CO2, r.Temperature, r.Humidity, r.Timestamp)
	if len(r.Tags) > 0 {
		fmt.Printf("tags: %v\n", r.Tags)
	}

	// --- milestone 2 以降 ---
	// TODO(milestone-2): time.NewTicker で定期実行、または chissoku 常駐で行ストリーム読み。
	// TODO(milestone-3): Mackerel API へ POST。
	// TODO(milestone-5): config.yaml を読み込む。

	return nil
}

func getenvDefault(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}
