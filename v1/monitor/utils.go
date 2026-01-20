package monitor

import (
	"errors"
	"math"
	"strconv"
	"strings"
)

type (
	CPUStats struct {
		User float64
		Sys  float64
	}
)

func parseCPUFromInfo(info string) (*CPUStats, error) {
	lines := strings.Split(info, "\n")
	var user, sys float64
	found := false

	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		valStr := strings.TrimSpace(parts[1])

		if key == "used_cpu_user" || key == "used_cpu_sys" {
			if val, err := strconv.ParseFloat(valStr, 64); err == nil {
				// Защита от NaN, Inf, отрицательных и слишком больших значений
				if !math.IsNaN(val) && !math.IsInf(val, 0) && val >= 0 && val < 1e9 {
					if key == "used_cpu_user" {
						user = val
					} else {
						sys = val
					}
					found = true
				}
			}
		}
	}

	if !found {
		return nil, errors.New("no valid CPU stats found in INFO")
	}

	return &CPUStats{User: user, Sys: sys}, nil
}
