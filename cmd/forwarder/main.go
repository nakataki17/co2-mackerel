// Command forwarder は chissoku の出力を Mackerel に送るフォワーダー（開発中）。
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/nakataki17/co2-mackerel/internal/chissoku"
	"github.com/nakataki17/co2-mackerel/internal/mackerel"
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
		DockerImage:       getenvDefault("CHISSOKU_DOCKER_IMAGE", chissoku.DefaultDockerImage),
		DockerExe:         os.Getenv("CHISSOKU_DOCKER_EXE"),
		Device:            getenvDefault("DEVICE", "/dev/ttyACM0"),
		StdoutIntervalSec: interval,
	}

	apiKey := os.Getenv("MACKEREL_API_KEY")
	serviceName := getenvDefault("MACKEREL_SERVICE_NAME", "environmental-sensors")
	metricsPrefix := getenvDefault("MACKEREL_METRICS_PREFIX", "co2.living")
	baseURL := getenvDefault("MACKEREL_API_BASE", mackerel.DefaultBaseURL)

	timeout := 30 * time.Second
	if s := os.Getenv("MACKEREL_TIMEOUT_SEC"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			timeout = time.Duration(v) * time.Second
		}
	}
	httpClient := &http.Client{Timeout: timeout}

	var skipMackerelOnce sync.Once

	return chissoku.StreamLines(ctx, opt, func(line []byte) error {
		r, err := reading.ParseLine(line)
		if err != nil {
			return err
		}
		printReading(r)
		if apiKey == "" {
			skipMackerelOnce.Do(func() {
				fmt.Fprintln(os.Stderr, "mackerel: MACKEREL_API_KEY is not set; skipping POST")
			})
			return nil
		}
		metrics, err := mackerel.BuildMetrics(metricsPrefix, r)
		if err != nil {
			return err
		}
		return mackerel.PostServiceMetrics(ctx, httpClient, baseURL, serviceName, apiKey, metrics)
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
