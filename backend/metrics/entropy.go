package metrics

import "math"

// TileEntropy calculates H = -∑ p_i log2(p_i)
func TileEntropy(grid [][]int) float64 {
	counts := make(map[int]int)
	total := 0
	for y := range grid {
		for x := range grid[y] {
			counts[grid[y][x]]++
			total++
		}
	}
	H := 0.0
	for _, cnt := range counts {
		p := float64(cnt) / float64(total)
		if p > 0 {
			H -= p * (math.Log2(p))
		}
	}
	return H
}
