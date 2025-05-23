package main

import (
	"DnD_Mapgenerator/backend/gol"
	"DnD_Mapgenerator/backend/metrics"
	"DnD_Mapgenerator/backend/mlca"
	"DnD_Mapgenerator/backend/noise"
	"DnD_Mapgenerator/backend/wfc"
	"encoding/base64"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Static("/", "frontend")
	e.Static("/maps", "saved_maps")

	e.POST("/save", saveMap)
	e.GET("/load", loadMap)
	e.POST("/generate", generateTiles)

	// Add a catch-all route for debugging purposes
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

	imageData := req.ImageData[strings.IndexByte(req.ImageData, ',')+1:]
	data, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		log.Printf("Invalid image data: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid image data")
	}

	fileName := fmt.Sprintf("map-%d.png", time.Now().Unix())
	if err = os.WriteFile("saved_maps/"+fileName, data, 0644); err != nil {
		log.Printf("Failed to save the image: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save the image")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Map saved", "fileName": fileName})
}

func loadMap(c echo.Context) error {
	var imageFiles []string

	err := filepath.Walk("saved_maps", func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			imageFiles = append(imageFiles, filepath.Base(path))
		}
		return nil
	})

	if err != nil {
		log.Printf("Failed to load map images: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load map images")
	}

	return c.JSON(http.StatusOK, imageFiles)
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
	Grid        [][]int             `json:"grid"`
	Entropy     float64             `json:"entropy"`
	Adjacency   map[int]map[int]int `json:"adjacency"`
	Clusters    []int               `json:"clusters"`
	Frequencies map[int]float64     `json:"frequencies"`
}

func generateTiles(c echo.Context) error {
	req := new(GenerateRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	switch req.GenerationMethod {

	case "mlca":
		// Create default rules
		rules := mlca.CreateDefaultRules()

		// Map paintedTiles ([][]int) to [][]mlca.TileColorType
		painted := make([][]mlca.TileColorType, len(req.PaintedTiles))
		for y := range req.PaintedTiles {
			painted[y] = make([]mlca.TileColorType, len(req.PaintedTiles[y]))
			for x, v := range req.PaintedTiles[y] {
				painted[y][x] = mlca.TileColorType(v)
			}
		}

		// Call with the converted grid
		grid, err := mlca.GenerateTiles(
			req.Width,
			req.Height,
			painted,
			req.Iterations,
			req.RandomnessFactor,
			rules,
		)
		if err != nil {
			log.Printf("Tile generation error: %v\n", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Tile generation failed",
			})
		}

		tileGrid := grid

		// convert to [][]int
		intGrid := make([][]int, len(tileGrid))
		for y := range tileGrid {
			intGrid[y] = make([]int, len(tileGrid[y]))
			for x := range tileGrid[y] {
				intGrid[y][x] = int(tileGrid[y][x].Color)
			}
		}

		ent := metrics.TileEntropy(intGrid)
		adj := metrics.AdjacencyMatrix(intGrid)
		clusters := metrics.ClusterSizes(intGrid)
		//fdim := metrics.FractalDimension(grid)
		//spec := metrics.SpectralSpectrum(grid)
		//auto := metrics.Autocorrelation(grid)
		//grad := metrics.GradientHistogram(grid)
		freq := metrics.TileFrequencies(intGrid)

		resp := GenerateResponse{
			Grid:        intGrid,
			Entropy:     ent,
			Adjacency:   adj,
			Clusters:    clusters,
			Frequencies: freq,
		}
		// Output
		return c.JSON(http.StatusOK, resp)

	case "noise":

		// Configure Noise-Generator
		seed := int64(1)
		ng := noise.NewNoiseGenerator(
			seed,
			req.NoiseScale,
			req.NoiseOctaves,
			req.NoisePersistence,
			req.NoiseLacunarity,
		)

		// Generate map
		tileGrid := ng.Generate(req.Width, req.Height)

		// Convert to [][]int
		intGrid := make([][]int, req.Height)
		for y := 0; y < req.Height; y++ {
			intGrid[y] = make([]int, req.Width)
			for x := 0; x < req.Width; x++ {
				intGrid[y][x] = int(tileGrid[y][x].Color)
			}
		}

		ent := metrics.TileEntropy(intGrid)
		adj := metrics.AdjacencyMatrix(intGrid)
		clusters := metrics.ClusterSizes(intGrid)
		freq := metrics.TileFrequencies(intGrid)

		resp := GenerateResponse{
			Grid:        intGrid,
			Entropy:     ent,
			Adjacency:   adj,
			Clusters:    clusters,
			Frequencies: freq,
		}
		// Output
		return c.JSON(http.StatusOK, resp)

	case "wfc":
		// Create a new grid
		gridObj := wfc.NewGrid(req.Width, req.Height)

		// Solve grid
		tiles, err := gridObj.Solve(100)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		// Convert [][]wfc.TileType to [][]int
		intGrid := make([][]int, len(tiles))
		for y := range tiles {
			intGrid[y] = make([]int, len(tiles[y]))
			for x := range tiles[y] {
				intGrid[y][x] = int(tiles[y][x])
			}
		}

		ent := metrics.TileEntropy(intGrid)
		adj := metrics.AdjacencyMatrix(intGrid)
		clusters := metrics.ClusterSizes(intGrid)
		freq := metrics.TileFrequencies(intGrid)

		resp := GenerateResponse{
			Grid:        intGrid,
			Entropy:     ent,
			Adjacency:   adj,
			Clusters:    clusters,
			Frequencies: freq,
		}
		// Output
		return c.JSON(http.StatusOK, resp)

	case "gol":
		// Reconstruct grid from prevGrid or create new
		var tileGrid [][]gol.Tile
		if len(req.PrevGrid) > 0 {
			// Reconstruct
			tileGrid = make([][]gol.Tile, req.Height)
			for y := 0; y < req.Height; y++ {
				tileGrid[y] = make([]gol.Tile, req.Width)
				for x := 0; x < req.Width; x++ {
					tileGrid[y][x].State = gol.TileState(req.PrevGrid[y][x])
				}
			}
		} else {
			// randomize new
			tileGrid = gol.NewGrid(req.Width, req.Height)
		}

		// Print painted cells
		for y := 0; y < len(req.PaintedTiles) && y < len(tileGrid); y++ {
			for x := 0; x < len(req.PaintedTiles[y]) && x < len(tileGrid[0]); x++ {
				if req.PaintedTiles[y][x] == gol.Alive {
					tileGrid[y][x].State = gol.Alive
				} else if req.PaintedTiles[y][x] == gol.Dead {
					tileGrid[y][x].State = gol.Dead
				}
			}
		}

		// Evolution
		nextGrid, err := gol.StepCA(tileGrid, req.Iterations)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "CA generation failed: "+err.Error())
		}

		// Output
		return c.JSON(http.StatusOK, gol.TilesToIntGrid(nextGrid))

	default:
		return echo.NewHTTPError(http.StatusBadRequest, "Unknown generation Method")
	}
}
