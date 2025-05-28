package metrics

import (
	"math"
)

// FractalDimension calculates the fractal dimension via box counting.
// Expects square grid (H==W), otherwise H will be used.
func FractalDimension(grid [][]int) float64 {
	N := len(grid)
	// Create scales as potentiates of 2
	var scales []int
	for s := 1; s <= N; s *= 2 {
		scales = append(scales, s)
	}

	var logN, logInv []float64
	for _, s := range scales {
		// how many boxes per axis?
		boxes := N / s
		count := 0
		for by := 0; by < boxes; by++ {
			for bx := 0; bx < boxes; bx++ {
				hit := false
				// check if there is at least one filled tile in this box
				for y := by * s; y < (by+1)*s; y++ {
					for x := bx * s; x < (bx+1)*s; x++ {
						if grid[y][x] != 0 {
							hit = true
							break
						}
					}
					if hit {
						break
					}
				}
				if hit {
					count++
				}
			}
		}
		if count > 0 {
			logN = append(logN, math.Log(float64(count)))
			logInv = append(logInv, math.Log(1.0/float64(s)))
		}
	}

	// lineare regression for log(N) ~ m * log(1/ε)
	n := float64(len(logN))
	var sumX, sumY, sumXY, sumXX float64
	for i := range logN {
		x := logInv[i]
		y := logN[i]
		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
	}
	// Incline m = (n Σxy – Σx Σy) / (n Σx² – (Σx)²)
	m := (n*sumXY - sumX*sumY) / (n*sumXX - sumX*sumX)
	return m
}
