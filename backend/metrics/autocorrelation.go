package metrics

// Autocorrelation calculates the spatial autocorrelation for lags 0…maxLag.
// Returns: map[Lag]→Correlation coefficient, Lag as [dx,dy].
func Autocorrelation(grid [][]int, maxLag int) map[[2]int]float64 {
	H := len(grid)
	W := len(grid[0])
	n := float64(H * W)

	// Average
	var sum float64
	for i := 0; i < H; i++ {
		for j := 0; j < W; j++ {
			sum += float64(grid[i][j])
		}
	}
	mu := sum / n

	// Variant sum
	var varSum float64
	for i := 0; i < H; i++ {
		for j := 0; j < W; j++ {
			d := float64(grid[i][j]) - mu
			varSum += d * d
		}
	}

	out := make(map[[2]int]float64)
	for dx := 0; dx <= maxLag; dx++ {
		for dy := 0; dy <= maxLag; dy++ {
			var covSum float64
			var count float64
			for x := 0; x+dx < H; x++ {
				for y := 0; y+dy < W; y++ {
					covSum += (float64(grid[x][y]) - mu) * (float64(grid[x+dx][y+dy]) - mu)
					count++
				}
			}
			key := [2]int{dx, dy}
			if varSum > 0 {
				out[key] = covSum / varSum
			} else {
				out[key] = 0
			}
		}
	}
	return out
}
