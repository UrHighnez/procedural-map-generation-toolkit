package mlca

import (
	"errors"
	"log"
	"math/rand"
	"procedural-map-generation-toolkit/backend/tiles"
)

type Tile struct {
	Color tiles.TileType
}

type ColorConditionFunc func(t Tile, neighbors []Tile, x, y int, grid [][]Tile, randomness float64) bool
type TerrainRule struct {
	SourceColor, TargetColor tiles.TileType
	NeighborTypes            []tiles.TileType
	MinCount, MaxCount       int
}

func (r TerrainRule) Condition(t Tile, neighbors []Tile, _ int, _ int, _ [][]Tile, randomness float64) bool {
	count := CountTilesByType(neighbors, r.NeighborTypes...)
	passCount := (r.MinCount < 0 || count >= r.MinCount) && (r.MaxCount < 0 || count <= r.MaxCount)
	if randomness == 0 {
		return t.Color == r.SourceColor && passCount
	}
	return t.Color == r.SourceColor && passCount && rand.Float64() < randomness
}

func CreateDefaultRules() []TerrainRule {
	// Define water, land and foliage categories
	waterTypes := []tiles.TileType{tiles.DeepWater, tiles.Water, tiles.CoastalWater}
	landTypes := []tiles.TileType{tiles.Forest, tiles.Bushes, tiles.Grass, tiles.Sand, tiles.WetSand}
	grass := []tiles.TileType{tiles.Grass}
	bushes := []tiles.TileType{tiles.Bushes}
	forest := []tiles.TileType{tiles.Forest}

	// Convert foliage adjacent to water into sand
	coastalCleanup := []TerrainRule{
		{tiles.Forest, tiles.Sand, waterTypes, 1, -1},
		{tiles.Bushes, tiles.Sand, waterTypes, 1, -1},
	}

	// Convert grass into sand and sand into wet sand adjacent to water
	beachRules := []TerrainRule{
		{tiles.Grass, tiles.Sand, waterTypes, 1, -1},
		{tiles.Sand, tiles.WetSand, waterTypes, 2, -1},
	}

	// Terrain transitions (erosion and sediment buildup)
	terrainRules := []TerrainRule{
		// Downgrade toward water
		{tiles.WetSand, tiles.CoastalWater, landTypes, -1, 4},
		{tiles.CoastalWater, tiles.Water, landTypes, -1, 2},
		{tiles.Water, tiles.DeepWater, landTypes, -1, 1},
		// Upgrade away from water
		{tiles.DeepWater, tiles.Water, landTypes, 1, -1},
		{tiles.Water, tiles.CoastalWater, landTypes, 2, -1},
		{tiles.CoastalWater, tiles.WetSand, landTypes, 5, -1},
		{tiles.WetSand, tiles.Sand, landTypes, 6, -1},
		{tiles.Sand, tiles.Grass, landTypes, 7, -1},
	}

	// Vegetation transitions
	foliageRules := []TerrainRule{
		// **Birth**
		{tiles.Grass, tiles.Bushes, grass, 8, 8},
		{tiles.Grass, tiles.Bushes, bushes, 2, 7},
		{tiles.Bushes, tiles.Forest, bushes, 8, 8},
		{tiles.Bushes, tiles.Forest, forest, 3, 6},
		// **Survival**
		{tiles.Bushes, tiles.Bushes, bushes, 2, 7},
		{tiles.Forest, tiles.Forest, forest, 3, 6},
		// **Dying**
		{tiles.Bushes, tiles.Grass, bushes, -1, 1},
		{tiles.Bushes, tiles.Grass, bushes, 8, 8},
		{tiles.Forest, tiles.Bushes, forest, -1, 2},
		{tiles.Forest, tiles.Bushes, forest, 7, 8},
	}

	// Combine in order: coastal cleanup → beaches → terrain → vegetation
	rules := append(coastalCleanup, beachRules...)
	rules = append(rules, terrainRules...)
	rules = append(rules, foliageRules...)
	return rules
}

func CountTilesByType(neighbors []Tile, types ...tiles.TileType) int {
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

func GenerateTiles(width, height int, paintedTiles [][]tiles.TileType, iterations int, initialRandomnessFactor float64,
	rules []TerrainRule) ([][]Tile, error) {

	if len(paintedTiles) != height || len(paintedTiles[0]) != width {
		return nil, errors.New("paintedTiles dimensions do not match provided dimensions")
	}

	grid := initializeGrid(width, height, paintedTiles)

	for i := 0; i < iterations; i++ {
		// Increase the decay rate
		//decay := float64(i*i) / float64(iterations*iterations)
		randomnessFactor := initialRandomnessFactor
		//* (1.0 - decay)

		grid = applyRules(grid, rules, width, height, randomnessFactor)

		log.Printf("Iteration %d complete", i)
	}

	return grid, nil
}

func initializeGrid(width, height int, paintedTiles [][]tiles.TileType) [][]Tile {
	grid := make([][]Tile, height)
	paintedTilesNum := 0
	randomTilesNum := 0
	for y := 0; y < height; y++ {
		grid[y] = make([]Tile, width)
		for x := 0; x < width; x++ {
			if paintedTiles[y][x] != -1 {
				grid[y][x] = Tile{Color: paintedTiles[y][x]}
				paintedTilesNum += 1
				//log.Printf("Initialized painted tile at (%d, %d) with color %d", x, y, paintedTiles[y][x])
			} else {
				randomColor := tiles.TileType(rand.Intn(8))
				grid[y][x] = Tile{Color: randomColor}
				randomTilesNum += 1
				//log.Printf("Initialized random tile at (%d, %d) with color %d", x, y, randomColor)
			}
		}
	}
	log.Printf("Initialized %d painted tiles and %d random tiles", paintedTilesNum, randomTilesNum)
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
					//log.Printf("Applied rule from %d to %d at (%d, %d)", rule.SourceColor, rule.TargetColor, x, y)
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

// getAdjacentTiles returns 8 neighbors including out-of-bounds as DeepWater
func getAdjacentTiles(grid [][]Tile, x, y, width, height int) []Tile {
	coordinates := []struct{ dx, dy int }{
		{-1, -1}, {-1, 0}, {-1, 1},
		{0, -1}, {0, 1},
		{1, -1}, {1, 0}, {1, 1},
	}
	neighbors := make([]Tile, 0, 8)
	for _, c := range coordinates {
		nx, ny := x+c.dx, y+c.dy
		if nx >= 0 && nx < width && ny >= 0 && ny < height {
			neighbors = append(neighbors, grid[ny][nx])
		} else {
			// außerhalb der Karte → als Wasser behandeln:
			neighbors = append(neighbors, Tile{Color: tiles.DeepWater})
		}
	}
	return neighbors
}
