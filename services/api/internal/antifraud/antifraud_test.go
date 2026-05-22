package antifraud

import "testing"

func TestCadenceFlag(t *testing.T) {
	cases := []struct {
		name        string
		avgMs       float64
		variance    float64
		wantFlagged bool
		wantReason  string
	}{
		{"normal run", 400, 50, false, ""},
		{"normal walk", 600, 80, false, ""},
		{"too fast", 200, 50, true, "cadence_too_fast"},
		{"too slow", 1200, 50, true, "cadence_too_slow"},
		{"too uniform (robot)", 400, 2, true, "cadence_too_uniform"},
		{"too chaotic (shake)", 400, 400, true, "cadence_too_chaotic"},
		{"zero period", 0, 50, true, "cadence_unknown"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			flagged, reason := CadenceFlag(tc.avgMs, tc.variance)
			if flagged != tc.wantFlagged {
				t.Fatalf("flagged = %v, want %v (reason=%q)", flagged, tc.wantFlagged, reason)
			}
			if reason != tc.wantReason {
				t.Fatalf("reason = %q, want %q", reason, tc.wantReason)
			}
		})
	}
}
