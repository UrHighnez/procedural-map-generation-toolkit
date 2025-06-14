package main

import (
	"encoding/base64"
	"fmt"
	"io/fs"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"procedural-map-generation-toolkit/backend/gol"
	"procedural-map-generation-toolkit/backend/metrics"
	"procedural-map-generation-toolkit/backend/mlca"
	"procedural-map-generation-toolkit/backend/noise"
	"procedural-map-generation-toolkit/backend/tiles"
	"procedural-map-generation-toolkit/backend/wfc"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const fixedSeed = 1
const defaultWFCSeed = 1

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Static("/", "frontend")
	e.Static("/maps", "saved_maps")

	e.POST("/save", saveMap)
	e.GET("/load", loadMap)
	e.GET("/colors", func(c echo.Context) error {
		return c.JSON(http.StatusOK, tiles.TileColors)
	})
	e.POST("/generate", generateTiles)

	e.GET("/*", func(c echo.Context) error {
		log.Printf("Requested file: %s", c.Request().URL.Path)
		return c.File(filepath.Join("frontend", c.Request().URL.Path))
	})

	if err := e.Start(":8000"); err != nil {
		log.Fatalf("Echo server startup failed: %v", err)
	}
}

func saveMap(c echo.Context) error {
	type SaveRequest struct {
		ImageData string `json:"imageData"`
	}

	req := new(SaveRequest)
	if err := c.Bind(req); err != nil {
		log.Printf("Invalid request format: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	dataURL := req.ImageData
	comma := strings.IndexByte(dataURL, ',')
	if comma >= 0 {
		dataURL = dataURL[comma+1:]
	}
	imgData, err := base64.StdEncoding.DecodeString(dataURL)
	if err != nil {
		log.Printf("Invalid image data: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid image data")
	}

	fileName := fmt.Sprintf("map-%d.png", time.Now().Unix())
	if err = os.WriteFile("saved_maps/"+fileName, imgData, 0644); err != nil {
		log.Printf("Failed to save the image: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save the image")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Map saved", "fileName": fileName})
}

func loadMap(c echo.Context) error {
	var files []string
	err := filepath.Walk("saved_maps", func(path string, info fs.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			files = append(files, filepath.Base(path))
		}
		return nil
	})
	if err != nil {
		log.Printf("Failed to load map images: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load map images")
	}
	return c.JSON(http.StatusOK, files)
}

type GenerateRequest struct {
	GenerationMethod string  `json:"generationMethod"`
	Width            int     `json:"width"`
	Height           int     `json:"height"`
	Iterations       int     `json:"iterations"`
	RandomnessFactor float64 `json:"randomnessFactor"`
	PrevGrid         [][]int `json:"prevGrid"`
	PaintedTiles     [][]int `json:"paintedTiles"`
	NoiseScale       float64 `json:"noiseScale"`
	NoiseOctaves     int     `json:"noiseOctaves"`
	NoisePersistence float64 `json:"noisePersistence"`
	NoiseLacunarity  float64 `json:"noiseLacunarity"`
	WFCSeed          *int64  `json:"wfcSeed,omitempty"`
}

type GenerateResponse struct {
	Grid        [][]int                    `json:"grid"`
	Colors      [tiles.NumTileTypes]string `json:"colors"`
	Entropy     float64                    `json:"entropy"`
	Adjacency   map[int]map[int]int        `json:"adjacency"`
	Frequencies map[int]float64            `json:"frequencies"`
	Autocorr    map[string]float64         `json:"autocorr"`
	FractalDim  float64                    `json:"fractalDim"`
	Spectrum    [][]float64                `json:"spectrum"`
}

func generateTiles(c echo.Context) error {
	req := new(GenerateRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	var (
		intGrid [][]int
		ent     float64
		adj     map[int]map[int]int
		freq    map[int]float64
		auto    map[[2]int]float64
		fd      float64
		spec    [][]float64

		genErr error
	)

	switch req.GenerationMethod {
	case "mlca":
		intGrid, ent, adj, freq, auto, fd, spec, genErr = runMLCA(req)
	case "noise":
		intGrid, ent, adj, freq, auto, fd, spec, genErr = runNoise(req)
	case "wfc":
		intGrid, ent, adj, freq, auto, fd, spec, genErr = runWFC(req)
	case "gol":
		intGrid, ent, adj, freq, auto, fd, spec, genErr = runGOL(req)
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "Unknown generation method")
	}

	if genErr != nil {
		log.Printf("Generation error: %v", genErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": genErr.Error()})
	}

	autoStr := make(map[string]float64, len(auto))
	for k, v := range auto {
		key := fmt.Sprintf("%d,%d", k[0], k[1])
		autoStr[key] = v
	}

	resp := GenerateResponse{
		Grid:        intGrid,
		Colors:      tiles.TileColors,
		Entropy:     ent,
		Adjacency:   adj,
		Frequencies: freq,
		Autocorr:    autoStr,
		FractalDim:  fd,
		Spectrum:    spec,
	}
	return c.JSON(http.StatusOK, resp)
}

func runMLCA(req *GenerateRequest) ([][]int, float64, map[int]map[int]int, map[int]float64, map[[2]int]float64, float64, [][]float64, error) {
	// Convert painted, ensuring correct dimensions for mlca.GenerateTiles
	painted := make([][]tiles.TileType, req.Height)
	for y := 0; y < req.Height; y++ {
		painted[y] = make([]tiles.TileType, req.Width)
		for x := 0; x < req.Width; x++ {
			// Default to -1 (not painted)
			painted[y][x] = -1
			if y < len(req.PaintedTiles) && x < len(req.PaintedTiles[y]) {
				if req.PaintedTiles[y][x] != -1 {
					painted[y][x] = tiles.TileType(req.PaintedTiles[y][x])
				}
			}
		}
	}

	// Generate
	tileGrid, err := mlca.GenerateTiles(req.Width, req.Height, painted, req.Iterations, req.RandomnessFactor, mlca.CreateDefaultRules(), rand.New(rand.NewSource(fixedSeed)))
	if err != nil {
		return nil, 0, nil, nil, nil, 0, nil, err
	}
	intGrid := make([][]int, len(tileGrid))
	for y := range tileGrid {
		intGrid[y] = make([]int, len(tileGrid[y]))
		for x := range tileGrid[y] {
			intGrid[y][x] = int(tileGrid[y][x].Color)
		}
	}
	// Metrics
	ent := metrics.TileEntropy(intGrid)
	adj := metrics.AdjacencyMatrix(intGrid)
	freq := metrics.TileFrequencies(intGrid)
	auto := metrics.Autocorrelation(intGrid, 5)
	fd := metrics.FractalDimension(intGrid)
	spec := metrics.SpectralSpectrum(intGrid)

	return intGrid, ent, adj, freq, auto, fd, spec, nil
}

func runNoise(req *GenerateRequest) ([][]int, float64, map[int]map[int]int, map[int]float64, map[[2]int]float64, float64, [][]float64, error) {
	ng := noise.NewNoiseGenerator(int64(fixedSeed), req.NoiseScale, req.NoiseOctaves, req.NoisePersistence, req.NoiseLacunarity)
	tileGrid := ng.Generate(req.Width, req.Height)
	intGrid := make([][]int, req.Height)
	for y := 0; y < req.Height; y++ {
		intGrid[y] = make([]int, req.Width)
		for x := 0; x < req.Width; x++ {
			intGrid[y][x] = int(tileGrid[y][x].Color)
		}
	}
	ent := metrics.TileEntropy(intGrid)
	adj := metrics.AdjacencyMatrix(intGrid)
	freq := metrics.TileFrequencies(intGrid)
	auto := metrics.Autocorrelation(intGrid, 5)
	fd := metrics.FractalDimension(intGrid)
	spec := metrics.SpectralSpectrum(intGrid)

	return intGrid, ent, adj, freq, auto, fd, spec, nil
}

func runWFC(req *GenerateRequest) ([][]int, float64, map[int]map[int]int, map[int]float64, map[[2]int]float64, float64, [][]float64, error) {
	gridObj := wfc.NewGrid(req.Width, req.Height)

	currentWFCSeed := int64(defaultWFCSeed)
	if req.WFCSeed != nil {
		currentWFCSeed = *req.WFCSeed
	}

	tilesOut, err := gridObj.Solve(100, currentWFCSeed)

	if err != nil {
		return nil, 0, nil, nil, nil, 0, nil, err
	}

	intGrid := make([][]int, len(tilesOut))
	for y := range tilesOut {
		intGrid[y] = make([]int, len(tilesOut[y]))
		for x := range tilesOut[y] {
			intGrid[y][x] = int(tilesOut[y][x])
		}
	}
	ent := metrics.TileEntropy(intGrid)
	adj := metrics.AdjacencyMatrix(intGrid)
	freq := metrics.TileFrequencies(intGrid)
	auto := metrics.Autocorrelation(intGrid, 5)
	fd := metrics.FractalDimension(intGrid)
	spec := metrics.SpectralSpectrum(intGrid)

	return intGrid, ent, adj, freq, auto, fd, spec, nil
}

func runGOL(req *GenerateRequest) ([][]int, float64, map[int]map[int]int, map[int]float64, map[[2]int]float64, float64, [][]float64, error) {
	var tileGrid [][]gol.Tile
	if len(req.PrevGrid) > 0 {
		tileGrid = make([][]gol.Tile, req.Height)
		for y := 0; y < req.Height; y++ {
			tileGrid[y] = make([]gol.Tile, req.Width)
			for x := 0; x < req.Width; x++ {
				tileGrid[y][x].State = tiles.TileType(req.PrevGrid[y][x])
			}
		}
	} else {
		tileGrid = gol.NewGrid(req.Width, req.Height)
	}
	// Paint overrides
	for y := range req.PaintedTiles {
		for x := range req.PaintedTiles[y] {
			v := req.PaintedTiles[y][x]
			if v < 0 || v >= int(tiles.NumTileTypes) {
				continue
			}
			t := tiles.TileType(v)
			switch t {
			case tiles.Bushes, tiles.Sand:
				tileGrid[y][x].State = t
			default:
				log.Printf("Ignoring painted tile %d at %d,%d", v, x, y)
			}
		}
	}
	next, err := gol.StepCA(tileGrid, req.Iterations)
	if err != nil {
		return nil, 0, nil, nil, nil, 0, nil, err
	}
	intGrid := gol.TilesToIntGrid(next)
	ent := metrics.TileEntropy(intGrid)
	adj := metrics.AdjacencyMatrix(intGrid)
	freq := metrics.TileFrequencies(intGrid)
	auto := metrics.Autocorrelation(intGrid, 5)
	fd := metrics.FractalDimension(intGrid)
	spec := metrics.SpectralSpectrum(intGrid)

	return intGrid, ent, adj, freq, auto, fd, spec, nil
}
