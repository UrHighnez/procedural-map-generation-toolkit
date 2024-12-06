package main

import (
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
	e.POST("/generate", collapseTiles)

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

func collapseTiles(c echo.Context) error {
	type GenerateRequest struct {
		Width        int                   `json:"width"`
		Height       int                   `json:"height"`
		PaintedTiles [][]wfc.TileColorType `json:"paintedTiles"`
		Iterations   int                   `json:"iterations"`
	}

	req := new(GenerateRequest)
	if err := c.Bind(req); err != nil {
		log.Printf("Invalid request format: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	width, height, paintedTiles := req.Width, req.Height, req.PaintedTiles
	fmt.Printf("paintedTiles dimensions: %d x %d (expected: %d x %d)\n", len(paintedTiles), len(paintedTiles[0]), height, width)

	rules := wfc.CreateDefaultRules()
	grid, err := wfc.CollapseTiles(width, height, paintedTiles, req.Iterations, rules)
	if err != nil {
		log.Printf("Tile generation error: %v\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Tile generation failed",
		})
	}

	return c.JSON(http.StatusOK, grid)
}
