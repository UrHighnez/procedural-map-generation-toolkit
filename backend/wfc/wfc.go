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
}

type TerrainRule struct {
	SourceColor TileColorType
	TargetColor TileColorType
	Condition   func(Tile, []Tile) bool
}

func CreateDefaultRules() []TerrainRule {
	return []TerrainRule{
		createRule(Land, CoastalWater, landToCoastalWaterCondition),
		createRule(Land, Grass, landToGrassCondition),
		createRule(Grass, Land, grassToLandCondition),
		createRule(Grass, Forest, grassToForestCondition),
		createRule(Forest, Grass, forestToGrassCondition),
		createRule(CoastalWater, Land, coastalWaterToLandCondition),
		createRule(CoastalWater, Water, coastalWaterToWaterCondition),
		createRule(Water, CoastalWater, waterToCoastalWaterCondition),
	}
}

func createRule(source, target TileColorType, condition func(Tile, []Tile) bool) TerrainRule {
	return TerrainRule{
		SourceColor: source,
		TargetColor: target,
		Condition:   condition,
	}
}

func landToCoastalWaterCondition(t Tile, neighbors []Tile) bool {
	landCount := CountTilesByType(neighbors, Land, Grass, Forest)
	return t.Color == Land && landCount <= 2 && rand.Float64() < 0.1
}

func landToGrassCondition(t Tile, neighbors []Tile) bool {
	landCount := CountTilesByType(neighbors, Land, Grass, Forest)
	return t.Color == Land && landCount > 5 && rand.Float64() < 0.75
}

func grassToLandCondition(t Tile, neighbors []Tile) bool {
	landCount := CountTilesByType(neighbors, Land, Grass, Forest)
	return t.Color == Grass && landCount < 8 && rand.Float64() < 0.75
}

func grassToForestCondition(t Tile, neighbors []Tile) bool {
	landCount := CountTilesByType(neighbors, Land, Grass, Forest)
	forestCount := CountTilesByType(neighbors, Forest)
	return t.Color == Grass && landCount > 5 && forestCount > 0 && forestCount <= 6 && rand.Float64() < 0.5
}

func forestToGrassCondition(t Tile, neighbors []Tile) bool {
	forestCount := CountTilesByType(neighbors, Forest)
	return t.Color == Forest && (forestCount <= 4 || forestCount > 4) && rand.Float64() < 0.4
}

func coastalWaterToLandCondition(t Tile, neighbors []Tile) bool {
	landCount := CountTilesByType(neighbors, Land)
	return t.Color == CoastalWater && landCount >= 6 && rand.Float64() < 0.25
}

func coastalWaterToWaterCondition(t Tile, neighbors []Tile) bool {
	landCount := CountTilesByType(neighbors, Land)
	return t.Color == CoastalWater && landCount < 2 && rand.Float64() < 0.2
}

func waterToCoastalWaterCondition(t Tile, neighbors []Tile) bool {
	landCount := CountTilesByType(neighbors, Land)
	return t.Color == Water && landCount > 0 && rand.Float64() < 0.3
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
		grid = applyRules(grid, rules, width, height)
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
			} else {
				grid[y][x] = Tile{Color: TileColorType(rand.Intn(5))}
			}
		}
	}
	return grid
}

func applyRules(grid [][]Tile, rules []TerrainRule, width, height int) [][]Tile {
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
	for _, coord := range coordinates {
		if coord.x >= 0 && coord.x < width && coord.y >= 0 && coord.y < height {
			neighbors = append(neighbors, grid[coord.y][coord.x])
		}
	}
	return neighbors
}
