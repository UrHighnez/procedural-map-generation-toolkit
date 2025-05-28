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
	DeepWater:    "#00507f", // 0 DeepWater
	Water:        "#1085bc", // 1 Water
	CoastalWater: "#3eb3e6", // 2 CoastalWater
	WetSand:      "#b59752", // 3 WetSand
	Sand:         "#ffd675", // 4 Sand
	Grass:        "#78e85b", // 5 Grass
	Bushes:       "#4caf32", // 6 Bushes
	Forest:       "#2c7519", // 7 Forest
}
