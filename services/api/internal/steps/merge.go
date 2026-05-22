package steps

// Merge ZMP vs Strava: lấy max; flag nếu chênh > 30%.
// Trả (mergedSteps, flagged, source).
//   source = "zmp" | "strava" | "merged"
func Merge(zmpSteps, stravaSteps int) (int, bool, string) {
	if zmpSteps == 0 && stravaSteps == 0 {
		return 0, false, "zmp"
	}
	if zmpSteps == 0 {
		return stravaSteps, false, "strava"
	}
	if stravaSteps == 0 {
		return zmpSteps, false, "zmp"
	}
	max := zmpSteps
	min := stravaSteps
	if stravaSteps > zmpSteps {
		max = stravaSteps
		min = zmpSteps
	}
	flagged := false
	if min > 0 && float64(max-min)/float64(max) > 0.30 {
		flagged = true
	}
	return max, flagged, "merged"
}
