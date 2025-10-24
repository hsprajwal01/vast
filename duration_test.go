package vast

import (
	"testing"
	"time"
)

func TestDurationUnmarshalText(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantSeconds float64
		wantError   bool
		description string
	}{
		// Valid standard durations
		{"Valid_5sec", "00:00:05", 5, false, "Valid 5 seconds"},
		{"Valid_10sec", "00:00:10", 10, false, "Valid 10 seconds"},
		{"Valid_15sec", "00:00:15", 15, false, "Valid 15 seconds"},
		{"Valid_25sec", "00:00:25", 25, false, "Valid 25 seconds"},
		{"Valid_30sec", "00:00:30", 30, false, "Valid 30 seconds"},
		{"Valid_40sec", "00:00:40", 40, false, "Valid 40 seconds"},
		{"Valid_50sec", "00:00:50", 50, false, "Valid 50 seconds"},
		{"Valid_1min", "00:01:00", 60, false, "Valid 1 minute"},
		{"Valid_2min30sec", "00:02:30", 150, false, "Valid 2:30"},
		{"Valid_1hour", "01:00:00", 3600, false, "Valid 1 hour"},
		{"Valid_2h30m45s", "02:30:45", 9045, false, "Valid 2:30:45"},
		{"Zero_duration", "00:00:00", 0, false, "Zero duration"},

		// Durations with milliseconds
		{"Valid_with_ms", "00:00:05.500", 5.5, false, "Duration with milliseconds"},

		// CRITICAL TEST CASES - These should now work with the fix
		{"Normalize_60sec", "00:00:60", 60, false, "60 seconds normalized to 1 minute"},
		{"Normalize_75sec", "00:00:75", 75, false, "75 seconds normalized to 1:15"},
		{"Normalize_60sec_with_ms", "00:00:60.123", 60.123, false, "60 seconds with milliseconds"},
		{"Normalize_65min", "00:65:00", 3900, false, "65 minutes normalized to 1:05:00"},
		{"Normalize_65min30sec", "00:65:30", 3930, false, "65:30 normalized to 1:05:30"},
		{"Normalize_120min", "00:120:00", 7200, false, "120 minutes normalized to 2 hours"},
		{"Normalize_3661sec", "00:00:3661", 3661, false, "3661 seconds normalized to 1:01:01"},

		// Edge cases
		{"Empty_string", "", 0, false, "Empty string returns zero"},
		{"Undefined", "undefined", 0, false, "Undefined returns zero"},
		{"Undefined_mixed_case", "UnDeFiNeD", 0, false, "Undefined (mixed case)"},
		{"Whitespace", "  00:00:05  ", 5, false, "Duration with whitespace"},

		// Invalid cases that should error
		{"Invalid_format_short", "00:00", 0, true, "Not enough parts"},
		{"Invalid_format_long", "00:00:00:00", 0, true, "Too many parts"},
		{"Invalid_non_numeric", "00:00:abc", 0, true, "Non-numeric seconds"},
		{"Invalid_negative", "00:00:-5", 0, true, "Negative value"},
		{"Invalid_ms_too_long", "00:00:00.9999", 0, true, "Invalid milliseconds (too many digits)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Duration
			err := d.UnmarshalText([]byte(tt.input))

			if tt.wantError {
				if err == nil {
					t.Errorf("%s: expected error but got none for input %q", tt.description, tt.input)
				} else {
					t.Logf("✓ PASS: %s - correctly raised error: %v", tt.description, err)
				}
				return
			}

			if err != nil {
				t.Errorf("%s: unexpected error for input %q: %v", tt.description, tt.input, err)
				return
			}

			gotSeconds := time.Duration(d).Seconds()
			if gotSeconds != tt.wantSeconds {
				t.Errorf("%s:\n  Input:    %q\n  Expected: %.3f seconds\n  Got:      %.3f seconds",
					tt.description, tt.input, tt.wantSeconds, gotSeconds)
			} else {
				t.Logf("✓ PASS: %s - Input: %q -> %.3f seconds", tt.description, tt.input, gotSeconds)
			}
		})
	}
}

// TestDurationMarshalText tests the formatting back to string
func TestDurationMarshalText(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"5_seconds", 5 * time.Second, "00:00:05"},
		{"1_minute", 60 * time.Second, "00:01:00"},
		{"75_seconds", 75 * time.Second, "00:01:15"},
		{"1_hour", 3600 * time.Second, "01:00:00"},
		{"2h30m45s", 2*time.Hour + 30*time.Minute + 45*time.Second, "02:30:45"},
		{"with_milliseconds", 5*time.Second + 500*time.Millisecond, "00:00:05.500"},
		{"zero", 0, "00:00:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Duration(tt.duration)
			got, err := d.MarshalText()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			gotStr := string(got)
			if gotStr != tt.want {
				t.Errorf("MarshalText():\n  Duration: %v\n  Expected: %q\n  Got:      %q",
					tt.duration, tt.want, gotStr)
			} else {
				t.Logf("✓ PASS: %v -> %q", tt.duration, gotStr)
			}
		})
	}
}

// TestDurationRoundTrip tests unmarshaling and marshaling together
func TestDurationRoundTrip(t *testing.T) {
	tests := []struct {
		input          string
		expectedOutput string // After normalization
	}{
		{"00:00:05", "00:00:05"},
		{"00:00:60", "00:01:00"}, // Normalized
		{"00:00:75", "00:01:15"}, // Normalized
		{"00:65:00", "01:05:00"}, // Normalized
		{"00:00:05.500", "00:00:05.500"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var d Duration

			// Unmarshal
			err := d.UnmarshalText([]byte(tt.input))
			if err != nil {
				t.Fatalf("UnmarshalText failed: %v", err)
			}

			// Marshal back
			got, err := d.MarshalText()
			if err != nil {
				t.Fatalf("MarshalText failed: %v", err)
			}

			gotStr := string(got)
			if gotStr != tt.expectedOutput {
				t.Errorf("Round trip failed:\n  Input:    %q\n  Expected: %q\n  Got:      %q",
					tt.input, tt.expectedOutput, gotStr)
			} else {
				t.Logf("✓ PASS: %q -> %q", tt.input, gotStr)
			}
		})
	}
}
