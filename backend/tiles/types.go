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

var TileColors = [NumTileTypes]string{
	DeepWater:    "#00507f",
	Water:        "#1085bc",
	CoastalWater: "#3eb3e6",
	WetSand:      "#b59752",
	Sand:         "#ffd675",
	Grass:        "#78e85b",
	Bushes:       "#4caf32",
	Forest:       "#2c7519",
}
