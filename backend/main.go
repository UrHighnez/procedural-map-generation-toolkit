package main

import (
	"DnD_Mapgenerator/backend/wfc"
	"encoding/base64"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Create a new Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Serve static frontend files
	e.Static("/", "frontend")
	// Serve saved map images
	e.Static("/maps", "saved_maps")

	// Routes for saving and loading maps
	e.POST("/save", saveMap)
	e.GET("/load", loadMap)

	// Route for generating tiles
	e.POST("/generate", generateTiles)

	// Start the Echo server
	err := e.Start(":8080")
	if err != nil {
		return
	}
}

func saveMap(c echo.Context) error {
	type SaveRequest struct {
		ImageData string `json:"imageData"`
	}

	req := new(SaveRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Decode base64 image data
	imageData := req.ImageData[strings.IndexByte(req.ImageData, ',')+1:]
	data, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid image data")
	}

	// Save the image as a file on the server
	fileName := fmt.Sprintf("map-%d.png", time.Now().Unix())
	err = os.WriteFile("saved_maps/"+fileName, data, 0644)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to save the image")
	}

	return c.JSON(http.StatusOK, "Map saved")
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
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to load map images")
	}

	return c.JSON(http.StatusOK, imageFiles)
}

func generateTiles(c echo.Context) error {
	type GenerateRequest struct {
		Width        int                   `json:"width"`
		Height       int                   `json:"height"`
		PaintedTiles [][]wfc.TileColorType `json:"paintedTiles"`
		Iterations   int                   `json:"iterations"`
	}

	req := new(GenerateRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	width := req.Width
	height := req.Height
	paintedTiles := req.PaintedTiles

	fmt.Printf("paintedTiles dimensions: %d x %d (expected: %d x %d)\n", len(paintedTiles), len(paintedTiles[0]), height, width)

	grid, err := wfc.GenerateTiles(width, height, paintedTiles, req.Iterations)
	if err != nil {
		fmt.Printf("Tile generation error: %v\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Tile generation failed",
		})
	}

	return c.JSON(http.StatusOK, grid)
}
