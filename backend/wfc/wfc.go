package wfc

import (
	"errors"
	"log"
	"math/rand"
)

type TileColorType int

const (
	Water TileColorType = iota
	CoastalWater
	Land
	Grass
	Forest
)

type Tile struct {
	Color TileColorType
}

type TerrainRule struct {
	SourceColor TileColorType
	TargetColor TileColorType
	Condition   func(Tile, []Tile, int, int, [][]Tile, float64) bool
}

func CreateDefaultRules() []TerrainRule {
	return []TerrainRule{
		createRule(Water, CoastalWater, waterToCoastalWaterCondition),
		createRule(CoastalWater, Water, coastalWaterToWaterCondition),
		createRule(CoastalWater, Land, coastalWaterToLandCondition),
		createRule(Land, CoastalWater, landToCoastalWaterCondition),
		createRule(Land, Grass, landToGrassCondition),
		createRule(Grass, Land, grassToLandCondition),
		createRule(Grass, Forest, grassToForestCondition),
		createRule(Forest, Grass, forestToGrassCondition),
	}
}

func createRule(source, target TileColorType, condition func(Tile, []Tile, int, int, [][]Tile, float64) bool) TerrainRule {
	return TerrainRule{
		SourceColor: source,
		TargetColor: target,
		Condition:   condition,
	}
}

func waterToCoastalWaterCondition(t Tile, neighbors []Tile, x, y int, grid [][]Tile, randomnessFactor float64) bool {
	landCount := CountTilesByType(neighbors, Land, Grass, Forest)
	return t.Color == Water && landCount > 0 && rand.Float64() < 0.3*randomnessFactor
}

func coastalWaterToWaterCondition(t Tile, neighbors []Tile, x, y int, grid [][]Tile, randomnessFactor float64) bool {
	landCount := CountTilesByType(neighbors, Land, Grass, Forest)
	return t.Color == CoastalWater && landCount < 3 && rand.Float64() < 0.4*randomnessFactor
}

func coastalWaterToLandCondition(t Tile, neighbors []Tile, x, y int, grid [][]Tile, randomnessFactor float64) bool {
	landCount := CountTilesByType(neighbors, Land, Grass, Forest)
	return t.Color == CoastalWater && landCount > 4 && rand.Float64() < 0.07*randomnessFactor
}

func landToCoastalWaterCondition(t Tile, neighbors []Tile, x, y int, grid [][]Tile, randomnessFactor float64) bool {
	landCount := CountTilesByType(neighbors, Land, Grass, Forest)
	return t.Color == Land && landCount < 4 && rand.Float64() < 0.6*randomnessFactor
}

func landToGrassCondition(t Tile, neighbors []Tile, x, y int, grid [][]Tile, randomnessFactor float64) bool {
	landCount := CountTilesByType(neighbors, Land, Grass, Forest)
	return t.Color == Land && landCount > 6 && rand.Float64() < 0.08*randomnessFactor
}

func grassToLandCondition(t Tile, neighbors []Tile, x, y int, grid [][]Tile, randomnessFactor float64) bool {
	landCount := CountTilesByType(neighbors, Land, Grass, Forest)
	return t.Color == Grass && landCount < 7 && rand.Float64() < 0.7*randomnessFactor
}

func grassToForestCondition(t Tile, neighbors []Tile, x, y int, grid [][]Tile, randomnessFactor float64) bool {
	landCount := CountTilesByType(neighbors, Land, Grass, Forest)
	forestCount := CountTilesByType(neighbors, Forest)
	return t.Color == Grass && landCount > 3 && forestCount > 1 && forestCount <= 4 && rand.Float64() < 0.4*randomnessFactor
}

func forestToGrassCondition(t Tile, neighbors []Tile, x, y int, grid [][]Tile, randomnessFactor float64) bool {
	forestCount := CountTilesByType(neighbors, Forest)
	if t.Color == Forest && forestCount >= 2 && forestCount <= 4 && rand.Float64() < 0.3*randomnessFactor {
		// Check for small clusters of forest tiles and break them up
		cluster := checkForSquareCluster(grid, x, y, Forest)
		return cluster
	}
	return t.Color == Forest && rand.Float64() < 0.2*randomnessFactor
}

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

	grid := initializeGrid(width, height, paintedTiles)

	for i := 0; i < iterations; i++ {
		// Increase the decay rate
		randomnessFactor := 1.0 - float64(i*i)/float64(iterations*iterations)

		log.Printf("Randomness factor at iteration %d: %f", i, randomnessFactor)

		grid = applyRules(grid, rules, width, height, randomnessFactor)

		log.Printf("Iteration %d complete", i)
	}

	return grid, nil
}

func initializeGrid(width, height int, paintedTiles [][]TileColorType) [][]Tile {
	grid := make([][]Tile, height)
	for y := 0; y < height; y++ {
		grid[y] = make([]Tile, width)
		for x := 0; x < width; x++ {
			if paintedTiles[y][x] != -1 {
				grid[y][x] = Tile{Color: paintedTiles[y][x]}
				log.Printf("Initialized painted tile at (%d, %d) with color %d", x, y, paintedTiles[y][x])
			} else {
				randomColor := TileColorType(rand.Intn(5))
				grid[y][x] = Tile{Color: randomColor}
				log.Printf("Initialized random tile at (%d, %d) with color %d", x, y, randomColor)
			}
		}
	}
	return grid
}

func applyRules(grid [][]Tile, rules []TerrainRule, width, height int, randomnessFactor float64) [][]Tile {
	nextGrid := make([][]Tile, height)
	for y := 0; y < height; y++ {
		nextGrid[y] = make([]Tile, width)
		for x := 0; x < width; x++ {
			currentTile := grid[y][x]
			neighbors := getAdjacentTiles(grid, x, y, width, height)
			ruleApplied := false
			for _, rule := range rules {
				if rule.Condition(currentTile, neighbors, x, y, grid, randomnessFactor) {
					nextGrid[y][x] = Tile{Color: rule.TargetColor}
					log.Printf("Applied rule from %d to %d at (%d, %d)", rule.SourceColor, rule.TargetColor, x, y)
					ruleApplied = true
					break
				}
			}
			if !ruleApplied {
				nextGrid[y][x] = currentTile
			}
		}
	}
	return nextGrid
}

type coordinate struct {
	x, y int
}

func getAdjacentTiles(grid [][]Tile, x, y, width, height int) []Tile {
	coordinates := []coordinate{
		{x - 1, y - 1},
		{x - 1, y},
		{x - 1, y + 1},
		{x, y - 1},
		{x, y + 1},
		{x + 1, y - 1},
		{x + 1, y},
		{x + 1, y + 1},
	}

	var neighbors []Tile
	for _, coords := range coordinates {
		if coords.x >= 0 && coords.x < width && coords.y >= 0 && coords.y < height {
			neighbors = append(neighbors, grid[coords.y][coords.x])
		}
	}
	return neighbors
}

func checkForSquareCluster(grid [][]Tile, x, y int, color TileColorType) bool {
	cluster := false
	// Add logic to check for a 2x2 cluster of the specified color
	// For example, check if (x,y), (x+1,y), (x,y+1), (x+1,y+1) are all of the specified color
	if (x+1 < len(grid[0]) && y+1 < len(grid)) &&
		grid[y][x].Color == color &&
		grid[y][x+1].Color == color &&
		grid[y+1][x].Color == color &&
		grid[y+1][x+1].Color == color {
		cluster = true
	}
	return cluster
}
