package wfc

import (
	"errors"
	"math/rand"
	"time"
)

type TileColorType int

const (
	DeepWater TileColorType = iota
	Water
	CoastalWater
	WetSand
	Sand
	Grass
	Bushes
	Forest
	NumTileTypes
)

// Rules: What tile-types can be adjacent
var adjacencyRules = map[TileColorType][]TileColorType{
	Forest:       {Bushes, Forest},
	Bushes:       {Sand, Grass, Bushes, Forest},
	Grass:        {WetSand, Sand, Grass, Bushes, Forest},
	Sand:         {CoastalWater, WetSand, Sand, Grass, Bushes},
	WetSand:      {Water, CoastalWater, WetSand, Sand, Grass},
	CoastalWater: {DeepWater, Water, CoastalWater, WetSand, Sand},
	Water:        {DeepWater, Water, CoastalWater},
	DeepWater:    {DeepWater, Water, CoastalWater},
}

// Define cell
type Cell struct {
	Possible  map[TileColorType]struct{} // Set von möglichen Tiles
	Collapsed bool
	Tile      TileColorType
}

// New wfc grid
func NewGrid(width, height int) [][]*Cell {
	grid := make([][]*Cell, height)
	for y := range grid {
		grid[y] = make([]*Cell, width)
		for x := range grid[y] {
			grid[y][x] = &Cell{
				Possible:  makeAllTileSet(),
				Collapsed: false,
			}
		}
	}
	return grid
}

// Helper function to make the tile set
func makeAllTileSet() map[TileColorType]struct{} {
	set := make(map[TileColorType]struct{})
	for t := TileColorType(0); t < NumTileTypes; t++ {
		set[t] = struct{}{}
	}
	return set
}

// Find the cell with the least uncertainty
func findMinEntropyCell(grid [][]*Cell) (x, y int, found bool) {
	minChoices := int(NumTileTypes) + 1
	minCells := [][2]int{}
	for row := range grid {
		for col := range grid[row] {
			cell := grid[row][col]
			if cell.Collapsed {
				continue
			}
			choices := len(cell.Possible)
			if choices > 0 {
				if choices < minChoices {
					minChoices = choices
					minCells = [][2]int{{col, row}}
				} else if choices == minChoices {
					minCells = append(minCells, [2]int{col, row})
				}
			}
		}
	}
	if len(minCells) > 0 {
		idx := rand.Intn(len(minCells))
		x, y, found = minCells[idx][0], minCells[idx][1], true
	}
	return
}

// Collapse randomly
func collapseCell(cell *Cell, rng *rand.Rand) {
	n := len(cell.Possible)
	ix := rng.Intn(n)
	i := 0
	for t := range cell.Possible {
		if i == ix {
			cell.Tile = t
			cell.Possible = map[TileColorType]struct{}{t: {}}
			cell.Collapsed = true
			return
		}
		i++
	}
}

// Iterate through the grid
func RunWFC(width, height int, maxSteps int) ([][]TileColorType, error) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	grid := NewGrid(width, height)

	for step := 0; step < width*height && step < maxSteps; step++ {
		x, y, found := findMinEntropyCell(grid)
		if !found {
			// Alles kollabiert
			break
		}
		cell := grid[y][x]
		if len(cell.Possible) == 0 {
			return nil, errors.New("WFC-Konflikt: Keine Tiles möglich")
		}
		collapseCell(cell, rng)
		propagate(grid, x, y)
	}
	// Result
	out := make([][]TileColorType, height)
	for y := 0; y < height; y++ {
		out[y] = make([]TileColorType, width)
		for x := 0; x < width; x++ {
			c := grid[y][x]
			if !c.Collapsed {
				// Take random value
				for t := range c.Possible {
					out[y][x] = t
					break
				}
			} else {
				out[y][x] = c.Tile
			}
		}
	}
	return out, nil
}

func propagate(grid [][]*Cell, x, y int) {
	width := len(grid[0])
	height := len(grid)
	queue := [][2]int{{x, y}}
	for len(queue) > 0 {
		cx, cy := queue[0][0], queue[0][1]
		queue = queue[1:]
		c := grid[cy][cx]
		// Check neighbors and reduce possibilities
		for _, d := range [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
			nx, ny := cx+d[0], cy+d[1]
			if nx < 0 || nx >= width || ny < 0 || ny >= height {
				continue
			}
			neighbor := grid[ny][nx]
			if neighbor.Collapsed {
				continue
			}
			before := len(neighbor.Possible)
			// Allow neighbors based on adjacency rules
			allowed := make(map[TileColorType]struct{})
			for t := range neighbor.Possible {
				for at := range adjacencyRules[c.Tile] {
					if t == adjacencyRules[c.Tile][at] {
						allowed[t] = struct{}{}
						break
					}
				}
			}
			if len(allowed) < before {
				neighbor.Possible = allowed
				queue = append(queue, [2]int{nx, ny})
			}
		}
	}
}
