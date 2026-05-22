package antifraud

import "time"

// Checker wraps CadenceFlag + giới hạn bước/chunk theo config.
type Checker struct {
	MaxStepsPerHour int     // 0 = không kiểm
	MinCadenceMs    int
	MaxCadenceMs    int
}

func NewChecker(maxStepsPerHour, minCadence, maxCadence int) *Checker {
	if minCadence == 0 {
		minCadence = 300
	}
	if maxCadence == 0 {
		maxCadence = 900
	}
	if maxStepsPerHour == 0 {
		maxStepsPerHour = 12000
	}
	return &Checker{MaxStepsPerHour: maxStepsPerHour, MinCadenceMs: minCadence, MaxCadenceMs: maxCadence}
}

// Check trả về (flagged, reasons) cho 1 chunk.
func (c *Checker) Check(steps int, cadenceMs int, start, end time.Time) (bool, []string) {
	reasons := []string{}
	dur := end.Sub(start)
	if dur < time.Second {
		dur = time.Second
	}
	stepsPerHour := float64(steps) / dur.Hours()
	if stepsPerHour > float64(c.MaxStepsPerHour) {
		reasons = append(reasons, "rate_exceeds_max")
	}
	if cadenceMs > 0 {
		if cadenceMs < c.MinCadenceMs {
			reasons = append(reasons, "cadence_too_fast")
		} else if cadenceMs > c.MaxCadenceMs {
			reasons = append(reasons, "cadence_too_slow")
		}
	}
	return len(reasons) > 0, reasons
}
