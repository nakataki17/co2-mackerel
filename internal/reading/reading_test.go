package reading

import (
	"testing"
)

func TestParseLine(t *testing.T) {
	const sample = `{"co2":1242,"humidity":31.3,"temperature":29.4,"tags":["Living"],"timestamp":"2023-02-01T20:50:51.240+09:00"}`

	r, err := ParseLine([]byte(sample))
	if err != nil {
		t.Fatal(err)
	}
	if r.CO2 != 1242 {
		t.Errorf("CO2: got %d want 1242", r.CO2)
	}
	if r.Temperature != 29.4 {
		t.Errorf("Temperature: got %v want 29.4", r.Temperature)
	}
	if r.Humidity != 31.3 {
		t.Errorf("Humidity: got %v want 31.3", r.Humidity)
	}
	if len(r.Tags) != 1 || r.Tags[0] != "Living" {
		t.Errorf("Tags: got %v want [Living]", r.Tags)
	}
	if r.Timestamp != "2023-02-01T20:50:51.240+09:00" {
		t.Errorf("Timestamp: got %q", r.Timestamp)
	}
}

func TestParseLine_empty(t *testing.T) {
	_, err := ParseLine([]byte("   \n"))
	if err == nil {
		t.Fatal("want error for empty line")
	}
}

func TestParseLine_invalidJSON(t *testing.T) {
	_, err := ParseLine([]byte(`not json`))
	if err == nil {
		t.Fatal("want error for invalid json")
	}
}
