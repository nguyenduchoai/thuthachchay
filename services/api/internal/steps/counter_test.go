package steps

import "testing"

func TestMerge(t *testing.T) {
	cases := []struct {
		name              string
		zmp, strava       int
		wantMerged        int
		wantFlagged       bool
	}{
		{"zmp only", 8000, 0, 8000, false},
		{"strava only", 0, 9000, 9000, false},
		{"close values", 9500, 10000, 10000, false},
		{"diverge > 30%", 5000, 10000, 10000, true},
		{"both zero", 0, 0, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			merged, flagged, _ := Merge(tc.zmp, tc.strava)
			if merged != tc.wantMerged {
				t.Errorf("merged = %d, want %d", merged, tc.wantMerged)
			}
			if flagged != tc.wantFlagged {
				t.Errorf("flagged = %v, want %v", flagged, tc.wantFlagged)
			}
		})
	}
}
