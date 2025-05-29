package metrics

// IsUniform checks if all values in a 2D integer grid are the same.
// It returns true if the grid is uniform (or empty), false otherwise.
func IsUniform(grid [][]int) bool {
	if len(grid) == 0 || len(grid[0]) == 0 {
		return true // An empty grid is technically uniform
	}

	firstValue := grid[0][0]
	for _, row := range grid {
		for _, cell := range row {
			if cell != firstValue {
				return false // Found a different value
			}
		}
	}
	return true // All values were the same
}
