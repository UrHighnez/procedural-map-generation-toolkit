package wfc

import (
	"errors"
	"log"
	"math/rand"
)

type TileColorType int

const (
	Land TileColorType = iota
	CoastalWater
	Water
	Grass
	Forest
)

type Tile struct {
	Color TileColorType
	//TerrainStrength float64
	//GrowthPotential float64
}

type TerrainRule struct {
	SourceColor TileColorType
	TargetColor TileColorType
	Condition   func(Tile, []Tile) bool
}

func CreateDefaultRules() []TerrainRule {
	return []TerrainRule{
		// Land to CoastalWater
		{
			SourceColor: Land,
			TargetColor: CoastalWater,
			Condition: func(t Tile, neighbors []Tile) bool {
				landCount := CountTilesByType(neighbors, Land, Grass, Forest)
				return t.Color == Land && landCount <= 1 && rand.Float64() < 0.1
			},
		},
		// Land to Grass
		{
			SourceColor: Land,
			TargetColor: Grass,
			Condition: func(t Tile, neighbors []Tile) bool {
				landCount := CountTilesByType(neighbors, Land, Grass, Forest)
				return t.Color == Land && landCount > 3 && rand.Float64() < 0.75
			},
		},
		// Grass to Land
		{
			SourceColor: Grass,
			TargetColor: Land,
			Condition: func(t Tile, neighbors []Tile) bool {
				landCount := CountTilesByType(neighbors, Land, Grass, Forest)
				return t.Color == Grass && landCount < 4 && rand.Float64() < 0.75
			},
		},
		// Grass to Forest
		{
			SourceColor: Grass,
			TargetColor: Forest,
			Condition: func(t Tile, neighbors []Tile) bool {
				landCount := CountTilesByType(neighbors, Land, Grass, Forest)
				forestCount := CountTilesByType(neighbors, Forest)
				return t.Color == Grass && landCount > 1 && forestCount > 0 && forestCount <= 3 && rand.Float64() < 0.3
			},
		},
		// Forest to Grass
		{
			SourceColor: Forest,
			TargetColor: Grass,
			Condition: func(t Tile, neighbors []Tile) bool {
				forestCount := CountTilesByType(neighbors, Forest)
				return t.Color == Forest && (forestCount <= 2 || forestCount > 3) && rand.Float64() < 0.4
			},
		},
		// CoastalWater to Land
		{
			SourceColor: CoastalWater,
			TargetColor: Land,
			Condition: func(t Tile, neighbors []Tile) bool {
				landCount := CountTilesByType(neighbors, Land)
				return t.Color == CoastalWater && landCount >= 3 && rand.Float64() < 0.25
			},
		},
		// CoastalWater to Water
		{
			SourceColor: CoastalWater,
			TargetColor: Water,
			Condition: func(t Tile, neighbors []Tile) bool {
				landCount := CountTilesByType(neighbors, Land)
				return t.Color == CoastalWater && landCount < 1 && rand.Float64() < 0.2
			},
		},
		// Water to CoastalWater
		{
			SourceColor: Water,
			TargetColor: CoastalWater,
			Condition: func(t Tile, neighbors []Tile) bool {
				landCount := CountTilesByType(neighbors, Land)
				return t.Color == Water && landCount > 0 && rand.Float64() < 0.3
			},
		},
	}
}

// CountTilesByType Helper function to count neighboring tiles of specified types
func CountTilesByType(neighbors []Tile, types ...TileColorType) int {
	count := 0
	for _, neighbor := range neighbors {
		for _, t := range types {
			if neighbor.Color == t {
				count++
				break
			}
		}
	}
	return count
}

func CollapseTiles(width, height int, paintedTiles [][]TileColorType, iterations int, rules []TerrainRule) ([][]Tile, error) {

	if len(paintedTiles) != height || len(paintedTiles[0]) != width {
		return nil, errors.New("paintedTiles dimensions do not match provided dimensions")
	}

	// Initialize the grid based on paintedTiles and random tiles where not specified
	grid := make([][]Tile, height)
	for y := 0; y < height; y++ {
		grid[y] = make([]Tile, width)
		for x := 0; x < width; x++ {
			if paintedTiles[y][x] != -1 {
				// Use the painted tile color if specified
				grid[y][x] = Tile{Color: paintedTiles[y][x]}
			} else {
				// Initialize with a random tile color otherwise
				grid[y][x] = Tile{Color: TileColorType(rand.Intn(5))} // Assuming 5 is the number of TileColorTypes
			}
		}
	}

	// Apply the constraints
	for i := 0; i < iterations; i++ {
		nextGrid := make([][]Tile, height)
		for y := 0; y < height; y++ {
			nextGrid[y] = make([]Tile, width)
			for x := 0; x < width; x++ {
				currentTile := grid[y][x]
				neighbors := getAdjacentTiles(grid, x, y, width, height)
				ruleApplied := false
				for _, rule := range rules {
					if rule.Condition(currentTile, neighbors) {
						nextGrid[y][x] = Tile{Color: rule.TargetColor}
						ruleApplied = true
						break
					}
				}
				if !ruleApplied {
					nextGrid[y][x] = currentTile // Retain the original tile if no rule is applied
				}
			}
		}
		grid = nextGrid
		log.Printf("Iteration %d complete", i)
	}
	return grid, nil
}

type coordinate struct {
	x, y int
}

func getAdjacentTiles(grid [][]Tile, x, y, width, height int) []Tile {
	coordinates := []coordinate{
		{x - 1, y},
		{x + 1, y},
		{x, y - 1},
		{x, y + 1},
	}

	var neighbors []Tile
	for _, coord := range coordinates {
		if coord.x >= 0 && coord.x < width && coord.y >= 0 && coord.y < height {
			neighbors = append(neighbors, grid[coord.y][coord.x])
		}
	}
	return neighbors
}
