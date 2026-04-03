// Package reading は chissoku の stdout に出る JSON 1 行を表す型とパース処理を置く。
package reading

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Reading は chissoku の JSON 1 行に対応する（docs/spec.md 3.1 参照）。
type Reading struct {
	CO2         int      `json:"co2"`
	Humidity    float64  `json:"humidity"`
	Temperature float64  `json:"temperature"`
	Tags        []string `json:"tags"`
	Timestamp   string   `json:"timestamp"`
}

// ParseLine は stdout から読んだ 1 行（JSON）を Reading に変換する。
func ParseLine(line []byte) (*Reading, error) {
	line = bytes.TrimSpace(line)
	if len(line) == 0 {
		return nil, fmt.Errorf("reading: empty line")
	}
	var r Reading
	if err := json.Unmarshal(line, &r); err != nil {
		return nil, fmt.Errorf("reading: parse json: %w", err)
	}
	return &r, nil
}
