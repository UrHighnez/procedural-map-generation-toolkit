package noise

import (
	"github.com/aquilax/go-perlin"
	"log"
	"procedural-map-generation-toolkit/backend/mlca"
	"procedural-map-generation-toolkit/backend/tiles"
)

// Generator generates a map based on Perlin-Noise
type Generator struct {
	perlin      *perlin.Perlin
	Scale       float64 // Scale the coordinates
	Octaves     int     // Number of Octaves
	Persistence float64 // Amplitude-degen per octave
	Lacunarity  float64 // Frequency-Multiplication per octave
	thresholds  []struct {
		Max   float64        // Upper limit of normalized noise-value
		Color tiles.TileType // Assign to TileColorType
	}
}

// NewNoiseGenerator creates a new NoiseGenerator with given parameters:
// Seed: For deterministic results
// Scale: Scale for noise-coordinates
// Octaves: Number of octaves
// Persistence: Amplitude-degeneration
// Lacunarity: Frequency-multiplication
func NewNoiseGenerator(seed int64, scale float64, octaves int, persistence, lacunarity float64) *Generator {
	// alpha = persistence, beta = lacunarity, n = octaves
	p := perlin.NewPerlin(persistence, lacunarity, int32(octaves), seed)

	// Define Thresholds for conversion of noise-values to tile-colors
	thresholds := []struct {
		Max   float64
		Color tiles.TileType
	}{
		{0.2, tiles.DeepWater},
		{0.4, tiles.Water},
		{0.5, tiles.CoastalWater},
		{0.55, tiles.WetSand},
		{0.6, tiles.Sand},
		{0.7, tiles.Grass},
		{0.8, tiles.Bushes},
		{1.0, tiles.Forest},
	}

	return &Generator{
		perlin:      p,
		Scale:       scale,
		Octaves:     octaves,
		Persistence: persistence,
		Lacunarity:  lacunarity,
		thresholds:  thresholds,
	}
}

// Generate a grid with width x height.
func (ng *Generator) Generate(width, height int) [][]mlca.Tile {

	minV, maxV := 1.0, 0.0
	var sumV float64
	var cnt int

	grid := make([][]mlca.Tile, height)
	for y := 0; y < height; y++ {
		grid[y] = make([]mlca.Tile, width)
		for x := 0; x < width; x++ {

			// Scale coordinates
			nx := float64(x) / float64(width) * ng.Scale
			ny := float64(y) / float64(height) * ng.Scale

			// Noise2D gives values between [-1,1]
			raw := ng.perlin.Noise2D(nx, ny)

			// Normalize to [0,1]
			normalized := (raw + 1) * 0.5

			if normalized < minV {
				minV = normalized
			}
			if normalized > maxV {
				maxV = normalized
			}
			sumV += normalized
			cnt++

			// Map to TileColorType
			color := ng.mapValueToColor(normalized)
			grid[y][x] = mlca.Tile{Color: color}
		}
	}

	meanV := sumV / float64(cnt)
	log.Printf(
		"Perlin-Noise @ scale=%.2f: min=%.3f max=%.3f mean=%.3f (samples=%d)",
		ng.Scale, minV, maxV, meanV, cnt,
	)
	return grid
}

// mapValueToColor maps a normalized noise-value to a tile-color
func (ng *Generator) mapValueToColor(val float64) tiles.TileType {
	for _, t := range ng.thresholds {
		if val <= t.Max {
			return t.Color
		}
	}
	return tiles.Forest
}
