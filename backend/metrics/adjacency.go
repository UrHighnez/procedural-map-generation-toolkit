package metrics

// AdjacencyMatrix[i][j] = Number of cell pair appearances i next to j
func AdjacencyMatrix(grid [][]int) map[int]map[int]int {
	adj := make(map[int]map[int]int)
	dirs := [][2]int{{1, 0}, {0, 1}, {-1, 0}, {0, -1}}
	H, W := len(grid), len(grid[0])
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			i := grid[y][x]
			if adj[i] == nil {
				adj[i] = make(map[int]int)
			}
			for _, d := range dirs {
				ny, nx := y+d[0], x+d[1]
				if ny >= 0 && ny < H && nx >= 0 && nx < W {
					j := grid[ny][nx]
					adj[i][j]++
				}
			}
		}
	}
	return adj
}
