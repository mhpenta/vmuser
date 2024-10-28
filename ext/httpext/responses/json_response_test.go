package responses

import (
	"encoding/json"
	"log/slog"
	"testing"
)

func TestDecodeUnicodeEscape(t *testing.T) {
	slog.Info("Running TestDecodeUnicodeEscape")

	data := `{
  "cik": "315189",
  "name": "DEERE \u0026 CO",
  "stateOfIncorporation": "DE",
  "stateLocation": "IL",
  "fiscalYearEnd": "1029",
  "businessPhone": "(309) 765-8000",
  "businessStreet": "ONE JOHN DEERE PLACE",
  "businessCity": "MOLINE",
  "businessState": "IL",
  "businessZip": "61265-8098",
  "tickers": [
    "DE"
  ],
  "exchange": "NYSE",
  "mostRecentFilingDate": "2023-10-03T14:53:54Z"
}`

	var company map[string]interface{}

	err := json.Unmarshal([]byte(data), &company)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	expectedName := "DEERE & CO"
	if company["name"] != expectedName {
		t.Fatalf("Expected company name to be %q but got %q", expectedName, company["name"])
	}
}
