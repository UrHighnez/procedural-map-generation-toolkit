package metrics

// TileFrequencies returns relative frequency per type
func TileFrequencies(grid [][]int) map[int]float64 {
	counts := make(map[int]int)
	total := 0
	for y := range grid {
		for x := range grid[y] {
			counts[grid[y][x]]++
			total++
		}
	}

	freq := make(map[int]float64, 8)

	for t := 0; t < 8; t++ {
		freq[t] = 0.0
	}

	for t, cnt := range counts {
		freq[t] = float64(cnt) / float64(total)
	}
	return freq
}
