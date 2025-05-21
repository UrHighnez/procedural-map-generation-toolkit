package metrics

// TileFrequencies liefert relative Häufigkeiten je Typ
func TileFrequencies(grid [][]int) map[int]float64 {
	counts := make(map[int]int)
	total := 0
	for y := range grid {
		for x := range grid[y] {
			counts[grid[y][x]]++
			total++
		}
	}
	freq := make(map[int]float64)
	for t, cnt := range counts {
		freq[t] = float64(cnt) / float64(total)
	}
	return freq
}
