package metrics

import (
	"math"
	"math/cmplx"
)

// SpectralSpectrum calculates the 2D DFT magnitude spectrum of the grid.
// Returns: Spectrum[u][v] = |Σ_x Σ_y grid[x][y]·e^(−2πi(ux/H+vy/W))|.
func SpectralSpectrum(grid [][]int) [][]float64 {
	H := len(grid)
	W := len(grid[0])
	spec := make([][]float64, H)
	for u := 0; u < H; u++ {
		spec[u] = make([]float64, W)
		for v := 0; v < W; v++ {
			var sum complex128
			for x := 0; x < H; x++ {
				for y := 0; y < W; y++ {
					angle := -2 * math.Pi * (float64(u*x)/float64(H) + float64(v*y)/float64(W))
					sum += complex(float64(grid[x][y]), 0) * cmplx.Exp(complex(0, angle))
				}
			}
			spec[u][v] = cmplx.Abs(sum)
		}
	}
	return spec
}
