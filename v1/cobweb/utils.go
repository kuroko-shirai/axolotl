package cobweb

import (
	"sort"
)

// median возвращает медиану списка вещественных чисел.
func median(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) * 0.5
	}

	return sorted[n/2]
}
