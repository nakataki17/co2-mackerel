// Package reading は chissoku の stdout に出る JSON 1 行を表す型とパース処理を置く。
package reading

import "fmt"

// Reading は chissoku の JSON 1 行に対応する（docs/spec.md 3.1 参照）。
type Reading struct {
	CO2         int       `json:"co2"`
	Humidity    float64   `json:"humidity"`
	Temperature float64   `json:"temperature"`
	Tags        []string  `json:"tags"`
	Timestamp   string    `json:"timestamp"`
}

// ParseLine は stdout から読んだ 1 行（JSON）を Reading に変換する。
//
// TODO(milestone-1): encoding/json の json.Unmarshal を使って実装する。
// ヒント: 空行や前後の空白は trim してから扱うとよい。
func ParseLine(line []byte) (*Reading, error) {
	return nil, fmt.Errorf("TODO(milestone-1): ParseLine を実装する")
}
