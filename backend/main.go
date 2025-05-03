package main

import (
	"DnD_Mapgenerator/backend/ca"
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
}

func generateTiles(c echo.Context) error {
	req := new(GenerateRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request")
	}

	switch req.GenerationMethod {

	case "wfc":
		// Create default rules
		rules := wfc.CreateDefaultRules()

		// Map paintedTiles ([][]int) to [][]wfc.TileColorType
		painted := make([][]wfc.TileColorType, len(req.PaintedTiles))
		for y := range req.PaintedTiles {
			painted[y] = make([]wfc.TileColorType, len(req.PaintedTiles[y]))
			for x, v := range req.PaintedTiles[y] {
				painted[y][x] = wfc.TileColorType(v)
			}
		}

		// Call with the converted grid
		grid, err := wfc.GenerateTiles(
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
		return c.JSON(http.StatusOK, grid)

	case "ca":
		// Reconstruct grid from prevGrid or create new
		var tileGrid [][]ca.Tile
		if len(req.PrevGrid) > 0 {
			// Reconstruct
			tileGrid = make([][]ca.Tile, req.Height)
			for y := 0; y < req.Height; y++ {
				tileGrid[y] = make([]ca.Tile, req.Width)
				for x := 0; x < req.Width; x++ {
					tileGrid[y][x].State = ca.TileState(req.PrevGrid[y][x])
				}
			}
		} else {
			// randomize new
			tileGrid = ca.NewGrid(req.Width, req.Height)
		}

		// Print painted cells
		for y := 0; y < len(req.PaintedTiles) && y < len(tileGrid); y++ {
			for x := 0; x < len(req.PaintedTiles[y]) && x < len(tileGrid[0]); x++ {
				if req.PaintedTiles[y][x] == 4 {
					tileGrid[y][x].State = ca.Alive
				} else if req.PaintedTiles[y][x] == 0 {
					tileGrid[y][x].State = ca.Dead
				}
			}
		}

		// Evolution
		nextGrid, err := ca.StepCA(tileGrid, req.Iterations)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "CA generation failed: "+err.Error())
		}

		// Response
		return c.JSON(http.StatusOK, ca.TilesToIntGrid(nextGrid))

	case "noise":
		return echo.NewHTTPError(http.StatusNotImplemented, "NOISE NOT IMPLEMENTED")

	default:
		return echo.NewHTTPError(http.StatusBadRequest, "Unknown generation Method")
	}
}
