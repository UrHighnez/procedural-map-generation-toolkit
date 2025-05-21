package wfc

import (
	"errors"
	"math/rand"
	"sort"
)

type TileType int

const (
	DeepWater TileType = iota
	Water
	CoastalWater
	WetSand
	Sand
	Grass
	Bushes
	Forest
	NumTileTypes
)

// adjacencyRules defines allowed neighbors for each tile.
var adjacencyRules = map[TileType][]TileType{
	DeepWater:    {DeepWater, Water, CoastalWater},
	Water:        {DeepWater, Water, CoastalWater, WetSand},
	CoastalWater: {DeepWater, Water, CoastalWater, WetSand, Sand},
	WetSand:      {Water, CoastalWater, WetSand, Sand, Grass},
	Sand:         {CoastalWater, WetSand, Sand, Grass, Bushes},
	Grass:        {WetSand, Sand, Grass, Bushes, Forest},
	Bushes:       {Sand, Grass, Bushes, Forest},
	Forest:       {Grass, Bushes, Forest},
}

type Cell struct {
	options   map[TileType]struct{} // remaining possible tiles
	tile      TileType              // collapsed tile
	collapsed bool
}

type Grid struct {
	width, height int
	cells         [][]*Cell
}

// NewGrid initializes a grid with all tiles possible in each cell.
func NewGrid(w, h int) *Grid {
	g := &Grid{width: w, height: h}
	g.cells = make([][]*Cell, h)
	for y := 0; y < h; y++ {
		g.cells[y] = make([]*Cell, w)
		for x := 0; x < w; x++ {
			// all tile types initially allowed
			opts := make(map[TileType]struct{}, NumTileTypes)
			for t := TileType(0); t < NumTileTypes; t++ {
				opts[t] = struct{}{}
			}
			g.cells[y][x] = &Cell{options: opts}
		}
	}
	return g
}

// Solve runs the WFC algorithm with a simple restart-on-conflict strategy.
func (g *Grid) Solve(maxRetries int) ([][]TileType, error) {

	const fixedSeed = 1
	rng := rand.New(rand.NewSource(fixedSeed))

	// Define water types
	waterSet := map[TileType]struct{}{DeepWater: {}, Water: {}, CoastalWater: {}}

	// Define land types
	landSet := map[TileType]struct{}{
		Sand: {}, Grass: {}, Bushes: {}, Forest: {},
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Cell Reset
		for y := 0; y < g.height; y++ {
			for x := 0; x < g.width; x++ {
				g.cells[y][x].collapsed = false
				g.cells[y][x].options = make(map[TileType]struct{}, NumTileTypes)
				for t := TileType(0); t < NumTileTypes; t++ {
					g.cells[y][x].options[t] = struct{}{}
				}
			}
		}
		// Set water on the border
		for y := 0; y < g.height; y++ {
			for x := 0; x < g.width; x++ {
				if x == 0 || y == 0 || x == g.width-1 || y == g.height-1 {
					opts := make(map[TileType]struct{}, len(waterSet))
					for t := range waterSet {
						opts[t] = struct{}{}
					}
					g.cells[y][x].options = opts
				}
			}
		}
		// Set a land circle in the center
		centerX, centerY := g.width/2, g.height/2
		const r = 3
		r2 := r * r
		for y := centerY - r; y <= centerY+r; y++ {
			if y < 0 || y >= g.height {
				continue
			}
			dy := y - centerY
			for x := centerX - r; x <= centerX+r; x++ {
				if x < 0 || x >= g.width {
					continue
				}
				dx := x - centerX
				if dx*dx+dy*dy <= r2 {
					opts := make(map[TileType]struct{}, len(landSet))
					for t := range landSet {
						opts[t] = struct{}{}
					}
					g.cells[y][x].options = opts
					// Collapse center
					//if dx == 0 && dy == 0 {
					//	choice := Forest
					//	g.cells[y][x].tile = choice
					//	g.cells[y][x].collapsed = true
					//}
				}
			}
		}
		// Collapse loop
		ok := true
		for {
			x, y, found := g.findMinEntropy(rng)
			if !found {
				// Check for conflict
				if g.anyCellHasNoOptions() {
					ok = false
				}
				break
			}
			if err := g.collapse(x, y, rng); err != nil {
				ok = false
				break
			}
			if err := g.propagate(); err != nil {
				ok = false
				break
			}
		}
		if ok {
			return g.export(), nil
		}
	}
	return nil, errors.New("WFC failed after retries")
}

// findMinEntropy picks a random cell with the fewest options (>1).
func (g *Grid) findMinEntropy(rng *rand.Rand) (int, int, bool) {
	minEntropy := NumTileTypes + 1
	var candidates [][2]int
	for y := 0; y < g.height; y++ {
		for x := 0; x < g.width; x++ {
			c := g.cells[y][x]
			if c.collapsed {
				continue
			}
			n := len(c.options)
			if n == 0 {
				return 0, 0, false // conflict
			}
			if n < int(minEntropy) {
				minEntropy = TileType(n)
				candidates = [][2]int{{x, y}}
			} else if n == int(minEntropy) {
				candidates = append(candidates, [2]int{x, y})
			}
		}
	}
	if len(candidates) == 0 {
		return 0, 0, false
	}
	i := rng.Intn(len(candidates))
	x, y := candidates[i][0], candidates[i][1]
	return x, y, true
}

// collapse chooses one deterministic option and marks the cell collapsed.
func (g *Grid) collapse(x, y int, rng *rand.Rand) error {
	c := g.cells[y][x]
	n := len(c.options)
	if n == 0 {
		return errors.New("conflict at collapse")
	}

	// Collect keys
	opts := make([]TileType, 0, n)
	for t := range c.options {
		opts = append(opts, t)
	}
	// Sort keys deterministically
	sort.Slice(opts, func(i, j int) bool { return opts[i] < opts[j] })

	// Create index via rng
	choice := opts[rng.Intn(n)]

	// Collapse cell
	c.options = map[TileType]struct{}{choice: {}}
	c.tile = choice
	c.collapsed = true
	return nil
}

// propagate enforces adjacency constraints across the grid.
func (g *Grid) propagate() error {
	queue := make([][2]int, 0)
	// initialize with all collapsed cells
	for y := 0; y < g.height; y++ {
		for x := 0; x < g.width; x++ {
			if g.cells[y][x].collapsed {
				queue = append(queue, [2]int{x, y})
			}
		}
	}
	// BFS
	dirs := [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	for len(queue) > 0 {
		x, y := queue[0][0], queue[0][1]
		queue = queue[1:]
		//base := g.cells[y][x]
		for _, d := range dirs {
			nx, ny := x+d[0], y+d[1]
			if nx < 0 || nx >= g.width || ny < 0 || ny >= g.height {
				continue
			}
			ne := g.cells[ny][nx]
			// collect allowed by all collapsed neighbors
			allowed := make(map[TileType]struct{})
			for t := TileType(0); t < NumTileTypes; t++ {
				allowed[t] = struct{}{}
			}
			for _, d2 := range dirs {
				x2, y2 := nx+d2[0], ny+d2[1]
				if x2 < 0 || x2 >= g.width || y2 < 0 || y2 >= g.height {
					continue
				}
				nbr := g.cells[y2][x2]
				if !nbr.collapsed {
					continue
				}
				tmp := make(map[TileType]struct{})
				for _, t2 := range adjacencyRules[nbr.tile] {
					if _, ok := allowed[t2]; ok {
						tmp[t2] = struct{}{}
					}
				}
				allowed = tmp
			}
			if len(allowed) == 0 {
				return errors.New("propagation conflict")
			}
			// if filtered, update and enqueue
			if len(allowed) < len(ne.options) {
				ne.options = allowed
				queue = append(queue, [2]int{nx, ny})
			}
		}
	}
	return nil
}

// export returns the final tile map once solved.
func (g *Grid) export() [][]TileType {
	out := make([][]TileType, g.height)
	for y := 0; y < g.height; y++ {
		out[y] = make([]TileType, g.width)
		for x := 0; x < g.width; x++ {
			out[y][x] = g.cells[y][x].tile
		}
	}
	return out
}

func (g *Grid) anyCellHasNoOptions() bool {
	for y := 0; y < g.height; y++ {
		for x := 0; x < g.width; x++ {
			if len(g.cells[y][x].options) == 0 {
				return true
			}
		}
	}
	return false
}
