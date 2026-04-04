// Package mackerel は Mackerel のサービスメトリクス API 向けの型とビルド処理を提供する。
package mackerel

import (
	"fmt"
	"time"

	"github.com/nakataki17/co2-mackerel/internal/reading"
)

// DefaultBaseURL は Mackerel API のベース URL。
const DefaultBaseURL = "https://api.mackerelio.com"

// ServiceMetric は POST /api/v0/services/{serviceName}/tsdb の 1 要素。
type ServiceMetric struct {
	Name  string  `json:"name"`
	Time  int64   `json:"time"`
	Value float64 `json:"value"`
}

// BuildMetrics は reading を 3 本のサービスメトリクスに変換する。
// prefix はメトリクス名の先頭（例: co2.living）。空なら co2.living を使う。
// 名前は {prefix}.ppm, {prefix}.temperature_c, {prefix}.humidity_pct（config.example.yaml の names に合わせる）。
func BuildMetrics(prefix string, r *reading.Reading) ([]ServiceMetric, error) {
	if r == nil {
		return nil, fmt.Errorf("mackerel: reading is nil")
	}
	pfx := prefix
	if pfx == "" {
		pfx = "co2.living"
	}
	ts, err := parseReadingTime(r.Timestamp)
	if err != nil {
		return nil, fmt.Errorf("mackerel: timestamp: %w", err)
	}
	t := ts.Unix()
	return []ServiceMetric{
		{Name: pfx + ".ppm", Time: t, Value: float64(r.CO2)},
		{Name: pfx + ".temperature_c", Time: t, Value: r.Temperature},
		{Name: pfx + ".humidity_pct", Time: t, Value: r.Humidity},
	}, nil
}

func parseReadingTime(s string) (time.Time, error) {
	if s == "" {
		return time.Now().UTC(), nil
	}
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, s)
}
