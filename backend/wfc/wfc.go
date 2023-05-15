package wfc

import (
	"errors"
	"math/rand"
)

type TileColorType int

const (
	Land TileColorType = iota
	CoastalWater
	Water
	Grass
)

type Tile struct {
	Color TileColorType
}

func GenerateTiles(width, height int, paintedTiles [][]TileColorType, iterations int) ([][]Tile, error) {

	if len(paintedTiles) != height {
		return nil, errors.New("paintedTiles height does not match the provided height")
	}

	for _, row := range paintedTiles {
		if len(row) != width {
			return nil, errors.New("paintedTiles width does not match the provided width")
		}
	}

	grid := make([][]Tile, height)
	for i := range grid {
		grid[i] = make([]Tile, width)
	}

	// Initialize the grid with random tiles
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if paintedTiles[y][x] != -1 {
				grid[y][x] = Tile{Color: paintedTiles[y][x]}
			} else {
				grid[y][x] = Tile{Color: TileColorType(rand.Intn(4))}
			}
		}
	}

	// Apply the constraints
	for i := 0; i < iterations; i++ { // Number of iterations
		nextGrid := make([][]Tile, height)
		for i := range nextGrid {
			nextGrid[i] = make([]Tile, width)
		}

		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				landCount := 0
				for _, adjacent := range adjacentCoordinates(x, y, width, height) {
					if grid[adjacent.y][adjacent.x].Color == Land || grid[adjacent.y][adjacent.x].Color == Grass {
						landCount++
					}
				}

				// Constraints
				if grid[y][x].Color == Land && (landCount <= 1) {
					if rand.Float64() < 0.5 { // 50% chance
						nextGrid[y][x] = Tile{Color: CoastalWater}
					} else {
						nextGrid[y][x] = grid[y][x]
					}
				} else if grid[y][x].Color == Land && (landCount >= 3) {
					if rand.Float64() < 0.5 { // 50% chance
						nextGrid[y][x] = Tile{Color: Grass}
					} else {
						nextGrid[y][x] = grid[y][x]
					}
				} else if grid[y][x].Color == Grass && (landCount < 3) {
					if rand.Float64() < 0.5 { // 50% chance
						nextGrid[y][x] = Tile{Color: Land}
					} else {
						nextGrid[y][x] = grid[y][x]
					}
				} else if grid[y][x].Color == CoastalWater && landCount >= 3 {
					if rand.Float64() < 0.5 { // 50% chance
						nextGrid[y][x] = Tile{Color: Land}
					} else {
						nextGrid[y][x] = grid[y][x]
					}
				} else if grid[y][x].Color == CoastalWater && landCount < 1 {
					if rand.Float64() < 0.5 { // 50% chance
						nextGrid[y][x] = Tile{Color: Water}
					} else {
						nextGrid[y][x] = grid[y][x]
					}
				} else if grid[y][x].Color == Water && landCount > 0 {
					if rand.Float64() < 0.5 { // 50% chance
						nextGrid[y][x] = Tile{Color: CoastalWater}
					} else {
						nextGrid[y][x] = grid[y][x]
					}
				} else {
					nextGrid[y][x] = grid[y][x]
				}
			}
		}

		grid = nextGrid
	}
	return grid, nil
}

type coordinate struct {
	x, y int
}

func adjacentCoordinates(x, y, width, height int) []coordinate {
	adjacent := []coordinate{
		{x - 1, y},
		{x + 1, y},
		{x, y - 1},
		{x, y + 1},
	}

	var validAdjacent []coordinate

	for _, coord := range adjacent {
		if coord.x >= 0 && coord.x < width && coord.y >= 0 && coord.y < height {
			validAdjacent = append(validAdjacent, coord)
		}
	}

	return validAdjacent
}
