package mackerel

import (
	"testing"

	"github.com/nakataki17/co2-mackerel/internal/reading"
)

func TestBuildMetrics(t *testing.T) {
	r := &reading.Reading{
		CO2:         1242,
		Humidity:    31.3,
		Temperature: 29.4,
		Timestamp:   "2023-02-01T20:50:51.240+09:00",
	}
	m, err := BuildMetrics("co2.living", r)
	if err != nil {
		t.Fatal(err)
	}
	if len(m) != 3 {
		t.Fatalf("len: %d", len(m))
	}
	want := []struct {
		name  string
		value float64
	}{
		{"co2.living.ppm", 1242},
		{"co2.living.temperature_c", 29.4},
		{"co2.living.humidity_pct", 31.3},
	}
	for i := range want {
		if m[i].Name != want[i].name {
			t.Errorf("[%d] name: got %q want %q", i, m[i].Name, want[i].name)
		}
		if m[i].Value != want[i].value {
			t.Errorf("[%d] value: got %v want %v", i, m[i].Value, want[i].value)
		}
		if m[i].Time == 0 {
			t.Errorf("[%d] time is zero", i)
		}
	}
}
