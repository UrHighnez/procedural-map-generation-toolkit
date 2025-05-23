package tiles

// TileType is the single source of truth for all tile‐color constants.
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
	NumTileTypes // for WFC loops
)
