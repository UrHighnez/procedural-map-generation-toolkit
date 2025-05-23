package gol

import (
	"errors"
	"math/rand"
	"procedural-map-generation-toolkit/backend/tiles"
)

const lifeProbability = 0.5

const alive = tiles.Bushes
const dead = tiles.Sand
const anyState = -1

type Tile struct {
	State tiles.TileType
}

type Rule struct {
	CurrentState       tiles.TileType
	TargetState        tiles.TileType
	NeighborState      tiles.TileType
	MinCount, MaxCount int
}

func (r Rule) Applies(t Tile, neighbors []Tile) bool {
	if r.CurrentState != anyState && t.State != r.CurrentState {
		return false
	}
	count := countNeighbors(neighbors, r.NeighborState)
	return count >= r.MinCount && (r.MaxCount < 0 || count <= r.MaxCount)
}

func InitializeGrid(width, height int) [][]Tile {
	grid := make([][]Tile, height)
	for y := 0; y < height; y++ {
		grid[y] = make([]Tile, width)
		for x := 0; x < width; x++ {
			if rand.Float64() < lifeProbability {

				// Cell Alive
				grid[y][x] = Tile{State: alive}
			} else {

				// Cell Dead
				grid[y][x] = Tile{State: dead}
			}
		}
	}
	return grid
}

func ApplyCARules(grid [][]Tile, rules []Rule, iterations int) ([][]Tile, error) {
	height := len(grid)
	if height == 0 {
		return nil, errors.New("grid height is zero")
	}
	width := len(grid[0])

	for i := 0; i < iterations; i++ {
		nextGrid := make([][]Tile, height)
		for y := 0; y < height; y++ {
			nextGrid[y] = make([]Tile, width)
			for x := 0; x < width; x++ {
				neighbors := getAdjacentTiles(grid, x, y, width, height)
				currentTile := grid[y][x]
				applied := false
				for _, rule := range rules {
					if rule.Applies(currentTile, neighbors) {
						nextGrid[y][x] = Tile{State: rule.TargetState}
						applied = true
						break
					}
				}
				if !applied {
					nextGrid[y][x] = currentTile
				}
			}
		}
		grid = nextGrid
	}

	return grid, nil
}

func countNeighbors(neighbors []Tile, targetState tiles.TileType) int {
	count := 0
	for _, n := range neighbors {
		if n.State == targetState {
			count++
		}
	}
	return count
}

func getAdjacentTiles(grid [][]Tile, x, y, width, height int) []Tile {
	dirs := []struct{ dx, dy int }{
		{-1, -1}, {-1, 0}, {-1, 1},
		{0, -1}, {0, 1},
		{1, -1}, {1, 0}, {1, 1},
	}
	var neighbors []Tile
	for _, d := range dirs {
		nx, ny := x+d.dx, y+d.dy
		if nx >= 0 && nx < width && ny >= 0 && ny < height {
			neighbors = append(neighbors, grid[ny][nx])
		}
	}
	return neighbors
}
func TilesToIntGrid(grid [][]Tile) [][]int {
	res := make([][]int, len(grid))
	for y := range grid {
		res[y] = make([]int, len(grid[0]))
		for x := range grid[0] {
			res[y][x] = int(grid[y][x].State)
		}
	}
	return res
}

// Rules for Conway’s Game of Life
func LifeRules() []Rule {
	return []Rule{

		// Survive with 2 or 3 living neighbors
		{CurrentState: alive, NeighborState: alive, MinCount: 2, MaxCount: 3, TargetState: alive},
		// Birth with 3 living neighbors
		{CurrentState: dead, NeighborState: alive, MinCount: 3, MaxCount: 3, TargetState: alive},
		// Under-/Overpopulation → DEATH
		{CurrentState: anyState, NeighborState: alive, MinCount: 0, MaxCount: 1, TargetState: dead},
		{CurrentState: anyState, NeighborState: alive, MinCount: 4, MaxCount: -1, TargetState: dead},
	}
}

func NewGrid(width, height int) [][]Tile {
	return InitializeGrid(width, height)
}

func StepCA(grid [][]Tile, iterations int) ([][]Tile, error) {
	return ApplyCARules(grid, LifeRules(), iterations)
}
