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
)

type Tile struct {
	Color TileColorType
}

func GenerateTiles(width, height int, paintedTiles [][]TileColorType) ([][]Tile, error) {

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
				grid[y][x] = Tile{Color: TileColorType(rand.Intn(3))}
			}
		}
	}

	// Apply the constraints
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if grid[y][x].Color == Land {
				// Check adjacent tiles
				for _, adjacent := range adjacentCoordinates(x, y, width, height) {
					if grid[adjacent.y][adjacent.x].Color == Water {
						grid[adjacent.y][adjacent.x].Color = CoastalWater
					}
				}
			}
		}
	}

	// Verify the constraints
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if grid[y][x].Color == Land {
				for _, adjacent := range adjacentCoordinates(x, y, width, height) {
					if grid[adjacent.y][adjacent.x].Color == Water {
						return nil, errors.New("constraint violation")
					}
				}
			}
		}
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

	validAdjacent := []coordinate{}

	for _, coord := range adjacent {
		if coord.x >= 0 && coord.x < width && coord.y >= 0 && coord.y < height {
			validAdjacent = append(validAdjacent, coord)
		}
	}

	return validAdjacent
}
