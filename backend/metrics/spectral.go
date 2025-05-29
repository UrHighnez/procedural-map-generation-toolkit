package metrics

import (
	"log"
	"math"
	"math/cmplx"
)

// SpectralSpectrum calculates the 2D DFT magnitude spectrum of the grid.
// Returns: Spectrum[u][v] = |Σ_x Σ_y grid[x][y]·e^(−2πi(ux/H+vy/W))|.
func SpectralSpectrum(grid [][]int) (spectrum [][]float64) { // Make sure it's exported
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC recovered in SpectralSpectrum: %v. Grid may be uniform or invalid. Details: %v", r, grid) // Log the grid
			// Ensure spectrum is initialized to a safe default if it's nil
			if len(grid) > 0 && len(grid[0]) > 0 {
				spectrum = make([][]float64, len(grid))
				for i := range spectrum {
					spectrum[i] = make([]float64, len(grid[0]))
				}
			} else { // Handle completely empty or malformed grid
				spectrum = make([][]float64, 0)
			}
		}
	}()

	if IsUniform(grid) {
		rows := len(grid)
		cols := 0
		if rows > 0 {
			cols = len(grid[0])
		}
		spectrum = make([][]float64, rows)
		for i := range spectrum {
			spectrum[i] = make([]float64, cols) // All zeros for a uniform grid
		}
		return spectrum
	}

	H := len(grid)
	W := len(grid[0])
	spectrum = make([][]float64, H)
	for u := 0; u < H; u++ {
		spectrum[u] = make([]float64, W)
		for v := 0; v < W; v++ {
			var sum complex128
			for x := 0; x < H; x++ {
				for y := 0; y < W; y++ {
					angle := -2 * math.Pi * (float64(u*x)/float64(H) + float64(v*y)/float64(W))
					sum += complex(float64(grid[x][y]), 0) * cmplx.Exp(complex(0, angle))
				}
			}
			spectrum[u][v] = cmplx.Abs(sum)
		}
	}
	return spectrum
}
