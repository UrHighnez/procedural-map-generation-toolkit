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
}

type GenerateResponse struct {
	Grid        [][]int                    `json:"grid"`
	Colors      [tiles.NumTileTypes]string `json:"colors"`
	Entropy     float64                    `json:"entropy"`
	Adjacency   map[int]map[int]int        `json:"adjacency"`
	Frequencies map[int]float64            `json:"frequencies"`
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

		genErr error
	)

	switch req.GenerationMethod {
	case "mlca":
		intGrid, ent, adj, freq, genErr = runMLCA(req)
	case "noise":
		intGrid, ent, adj, freq, genErr = runNoise(req)
	case "wfc":
		intGrid, ent, adj, freq, genErr = runWFC(req)
	case "gol":
		intGrid, ent, adj, freq, genErr = runGOL(req)
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "Unknown generation method")
	}

	if genErr != nil {
		log.Printf("Generation error: %v", genErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": genErr.Error()})
	}

	resp := GenerateResponse{
		Grid:        intGrid,
		Colors:      tiles.TileColors,
		Entropy:     ent,
		Adjacency:   adj,
		Frequencies: freq,
	}
	return c.JSON(http.StatusOK, resp)
}

func runMLCA(req *GenerateRequest) ([][]int, float64, map[int]map[int]int, map[int]float64, error) {
	// Convert painted
	painted := make([][]tiles.TileType, len(req.PaintedTiles))
	for y := range req.PaintedTiles {
		painted[y] = make([]tiles.TileType, len(req.PaintedTiles[y]))
		for x, v := range req.PaintedTiles[y] {
			painted[y][x] = tiles.TileType(v)
		}
	}
	// Generate
	tileGrid, err := mlca.GenerateTiles(req.Width, req.Height, painted, req.Iterations, req.RandomnessFactor, mlca.CreateDefaultRules(), rand.New(rand.NewSource(fixedSeed)))
	if err != nil {
		return nil, 0, nil, nil, err
	}
	// to ints
	intGrid := make([][]int, len(tileGrid))
	for y := range tileGrid {
		intGrid[y] = make([]int, len(tileGrid[y]))
		for x := range tileGrid[y] {
			intGrid[y][x] = int(tileGrid[y][x].Color)
		}
	}
	// metrics
	ent := metrics.TileEntropy(intGrid)
	adj := metrics.AdjacencyMatrix(intGrid)
	freq := metrics.TileFrequencies(intGrid)
	return intGrid, ent, adj, freq, nil
}

func runNoise(req *GenerateRequest) ([][]int, float64, map[int]map[int]int, map[int]float64, error) {
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
	return intGrid, ent, adj, freq, nil
}

func runWFC(req *GenerateRequest) ([][]int, float64, map[int]map[int]int, map[int]float64, error) {
	gridObj := wfc.NewGrid(req.Width, req.Height)
	tilesOut, err := gridObj.Solve(100)
	if err != nil {
		return nil, 0, nil, nil, err
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
	return intGrid, ent, adj, freq, nil
}

func runGOL(req *GenerateRequest) ([][]int, float64, map[int]map[int]int, map[int]float64, error) {
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
			switch tiles.TileType(req.PaintedTiles[y][x]) {
			case tiles.Bushes:
				tileGrid[y][x].State = tiles.Bushes
			case tiles.Sand:
				tileGrid[y][x].State = tiles.Sand
			default:
				panic("No valid tile type.")
			}
		}
	}
	next, err := gol.StepCA(tileGrid, req.Iterations)
	if err != nil {
		return nil, 0, nil, nil, err
	}
	intGrid := gol.TilesToIntGrid(next)
	ent := metrics.TileEntropy(intGrid)
	adj := metrics.AdjacencyMatrix(intGrid)
	freq := metrics.TileFrequencies(intGrid)
	return intGrid, ent, adj, freq, nil
}
