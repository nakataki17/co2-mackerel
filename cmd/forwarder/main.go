// Command forwarder は chissoku の出力を Mackerel に送るフォワーダー（開発中）。
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"

	"github.com/nakataki17/co2-mackerel/internal/chissoku"
	"github.com/nakataki17/co2-mackerel/internal/reading"
)

func main() {
	sigs := []os.Signal{os.Interrupt}
	if runtime.GOOS != "windows" {
		sigs = append(sigs, syscall.SIGTERM)
	}
	ctx, stop := signal.NotifyContext(context.Background(), sigs...)
	defer stop()

	if err := run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	interval := 60
	if s := os.Getenv("CHISSOKU_INTERVAL_SEC"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			interval = v
		}
	}

	opt := chissoku.RunOptions{
		BinPath:           os.Getenv("CHISSOKU_BIN"),
		Device:            getenvDefault("DEVICE", "/dev/ttyACM0"),
		StdoutIntervalSec: interval,
	}

	return chissoku.StreamLines(ctx, opt, func(line []byte) error {
		r, err := reading.ParseLine(line)
		if err != nil {
			return err
		}
		printReading(r)
		return nil
	})
}

func printReading(r *reading.Reading) {
	fmt.Printf("CO2=%d ppm, 温度=%.1f°C, 湿度=%.1f%%, 時刻=%s\n",
		r.CO2, r.Temperature, r.Humidity, r.Timestamp)
	if len(r.Tags) > 0 {
		fmt.Printf("tags: %v\n", r.Tags)
	}
}

func getenvDefault(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}
