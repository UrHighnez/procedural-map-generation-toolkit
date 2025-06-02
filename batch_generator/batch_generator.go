package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	baseURL       = "http://localhost:8000/generate"
	outputDir     = "output_maps"
	mapsPerMethod = 100
	defaultWidth  = 25
	defaultHeight = 25
)

// GenerateRequest matches the structure expected by the backend
type GenerateRequest struct {
	GenerationMethod string  `json:"generationMethod"`
	Width            int     `json:"width"`
	Height           int     `json:"height"`
	Iterations       int     `json:"iterations,omitempty"`
	RandomnessFactor float64 `json:"randomnessFactor,omitempty"`
	PaintedTiles     [][]int `json:"paintedTiles,omitempty"`
	NoiseScale       float64 `json:"noiseScale,omitempty"`
	NoiseOctaves     int     `json:"noiseOctaves,omitempty"`
	NoisePersistence float64 `json:"noisePersistence,omitempty"`
	NoiseLacunarity  float64 `json:"noiseLacunarity,omitempty"`
	WFCSeed          *int64  `json:"wfcSeed,omitempty"` // Pointer to allow omission
}

// GenerateResponse matches the structure returned by the backend
type GenerateResponse struct {
	Grid        [][]int                   `json:"grid"`
	Colors      []string                  `json:"colors"`
	Entropy     float64                   `json:"entropy"`
	Adjacency   map[string]map[string]int `json:"adjacency"`
	Frequencies map[string]float64        `json:"frequencies"`
	Autocorr    map[string]float64        `json:"autocorr"`
	FractalDim  float64                   `json:"fractalDim"`
	Spectrum    [][]float64               `json:"spectrum"`
}

// ResultData is saved for each map
type ResultData struct {
	RequestParams    GenerateRequest  `json:"requestParams"`
	ResponseMetrics  GenerateResponse `json:"responseMetrics"`
	GenerationTimeMs int64            `json:"generationTimeMs"`
}

func main() {
	methods := []string{"mlca", "noise", "wfc"}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create base output directory: %v", err)
	}

	for _, method := range methods {
		log.Printf("Starting generation for method: %s", method)
		methodDir := filepath.Join(outputDir, method)
		if err := os.MkdirAll(methodDir, 0755); err != nil {
			log.Fatalf("Failed to create output directory for %s: %v", method, err)
		}

		generatedCount := 0
		switch method {
		case "mlca":
			// 10 randomnessFactor values * 10 iteration values = 100 maps
			for i := 0; i < 10; i++ { // Iterations 1 to 10
				iterations := i + 1
				for j := 0; j < 10; j++ { // RandomnessFactor 0.0 to 0.9
					if generatedCount >= mapsPerMethod {
						break
					}
					randomness := float64(j) * 0.1
					params := GenerateRequest{
						GenerationMethod: method,
						Width:            defaultWidth,
						Height:           defaultHeight,
						Iterations:       iterations,
						RandomnessFactor: randomness,
						PaintedTiles:     [][]int{},
					}
					filename := fmt.Sprintf("mlca_iter_%d_rand_%.2f.json", iterations, randomness)
					generateAndSave(params, filepath.Join(methodDir, filename))
					generatedCount++
				}
				if generatedCount >= mapsPerMethod {
					break
				}
			}
		case "noise":
			// 10 noiseScale values * 10 noiseOctaves values = 100 maps
			octaveSteps := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

			// Default values
			defaultPersistence := 0.9
			defaultLacunarity := 1.8

			for i := 0; i < 10; i++ { // noiseScale from 0.2 to 2.0
				scale := 0.2 + (float64(i) * 0.2)
				for j := 0; j < 10; j++ { // noiseOctaves
					if generatedCount >= mapsPerMethod {
						break
					}
					octaves := octaveSteps[j]
					params := GenerateRequest{
						GenerationMethod: method,
						Width:            defaultWidth,
						Height:           defaultHeight,
						NoiseScale:       scale,
						NoiseOctaves:     octaves,
						NoisePersistence: defaultPersistence,
						NoiseLacunarity:  defaultLacunarity,
						PaintedTiles:     [][]int{},
					}
					filename := fmt.Sprintf("noise_scale_%.2f_oct_%d.json", scale, octaves)
					generateAndSave(params, filepath.Join(methodDir, filename))
					generatedCount++
				}
				if generatedCount >= mapsPerMethod {
					break
				}
			}
		case "wfc":
			for i := 0; i < mapsPerMethod; i++ {
				seed := int64(i + 1) // WFCSeed 1 to 100
				params := GenerateRequest{
					GenerationMethod: method,
					Width:            defaultWidth,
					Height:           defaultHeight,
					WFCSeed:          &seed, // Pass the pointer to the seed
					PaintedTiles:     [][]int{},
				}
				filename := fmt.Sprintf("wfc_seed_%d.json", seed)
				generateAndSave(params, filepath.Join(methodDir, filename))
				generatedCount++
			}
		}
		log.Printf("Finished generation for method: %s, %d maps generated.", method, generatedCount)
	}
	log.Println("All generations complete.")
}

func generateAndSave(params GenerateRequest, outputPath string) {
	log.Printf("Requesting: %s, Params: %+v", params.GenerationMethod, params)

	startTime := time.Now()

	jsonData, err := json.Marshal(params)
	if err != nil {
		log.Printf("Error marshalling request for %s: %v", outputPath, err)
		return
	}

	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error creating request for %s: %v", outputPath, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: time.Second * 60} // 60-second timeout
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending request for %s: %v", outputPath, err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	generationTimeMs := time.Since(startTime).Milliseconds()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("Error response for %s (Status %d): %s", outputPath, resp.StatusCode, string(bodyBytes))
		return
	}

	var genResponse GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResponse); err != nil {
		log.Printf("Error decoding response for %s: %v", outputPath, err)
		return
	}

	result := ResultData{
		RequestParams:    params,
		ResponseMetrics:  genResponse,
		GenerationTimeMs: generationTimeMs,
	}

	fileData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Printf("Error marshalling result for %s: %v", outputPath, err)
		return
	}

	if err := os.WriteFile(outputPath, fileData, 0644); err != nil {
		log.Printf("Error writing result file %s: %v", outputPath, err)
		return
	}
	log.Printf("Successfully generated and saved: %s (Time: %dms)", outputPath, generationTimeMs)
	time.Sleep(100 * time.Millisecond) // Brief pause to avoid overwhelming the server
}
