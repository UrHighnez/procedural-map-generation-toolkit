package metrics

// ClusterSizes returns slice with all cluster sizes
func ClusterSizes(grid [][]int) []int {
	H, W := len(grid), len(grid[0])
	seen := make([][]bool, H)
	for i := range seen {
		seen[i] = make([]bool, W)
	}
	var sizes []int
	dirs := [][2]int{{1, 0}, {0, 1}, {-1, 0}, {0, -1}}
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			if seen[y][x] {
				continue
			}
			t := grid[y][x]
			// BFS flood-fill
			queue := [][2]int{{y, x}}
			seen[y][x] = true
			size := 0
			for len(queue) > 0 {
				cy, cx := queue[0][0], queue[0][1]
				queue = queue[1:]
				size++
				for _, d := range dirs {
					ny, nx := cy+d[0], cx+d[1]
					if ny >= 0 && ny < H && nx >= 0 && nx < W && !seen[ny][nx] && grid[ny][nx] == t {
						seen[ny][nx] = true
						queue = append(queue, [2]int{ny, nx})
					}
				}
			}
			sizes = append(sizes, size)
		}
	}
	return sizes
}
