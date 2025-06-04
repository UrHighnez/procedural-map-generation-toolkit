package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	outputDir = "output_maps" // Match the output directory of the generation script
)

// GenerateRequest structure from the generation script
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
	WFCSeed          *int64  `json:"wfcSeed,omitempty"`
}

// GenerateResponse structure from the generation script
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

// ResultData structure from the generation script
type ResultData struct {
	RequestParams    GenerateRequest  `json:"requestParams"`
	ResponseMetrics  GenerateResponse `json:"responseMetrics"`
	GenerationTimeMs int64            `json:"generationTimeMs"`
	FilePath         string           `json:"-"` // Internal, not from JSON
}

// DataPoint for the CSV output of individual records
type DataPoint struct {
	FilePath                string  // To trace back to original data
	Method                  string  // Generation method
	Entropy                 float64 // Directly from metrics
	FractalDim              float64 // Directly from metrics
	LandRatio               float64 // Derived from Frequencies
	NumUniqueAdjacencyPairs int     // Derived from Adjacency
	AvgAutocorrLag1         float64 // Derived from Autocorr
	LowFreqEnergyRatio      float64 // Derived from Spectrum
}

// MethodAverageData holds the averaged metrics for a specific generation method
type MethodAverageData struct {
	Method                string
	AvgEntropy            float64
	AvgFractalDim         float64
	AvgLandRatio          float64
	AvgNumUniqueAdjPairs  float64
	AvgAvgAutocorrLag1    float64
	AvgLowFreqEnergyRatio float64
	SampleCount           int // Number of samples used for averaging
}

func main() {
	methods := []string{"mlca", "noise", "wfc"}
	var allResults []ResultData

	log.Println("Starting to read generated map data...")

	for _, method := range methods {
		methodDir := filepath.Join(outputDir, method)
		log.Printf("Processing directory: %s", methodDir)

		err := filepath.WalkDir(methodDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				log.Printf("Error accessing path %s: %v. Skipping.", path, err)
				return err
			}
			if !d.IsDir() && strings.HasSuffix(d.Name(), ".json") {
				fileData, readErr := os.ReadFile(path)
				if readErr != nil {
					log.Printf("Error reading file %s: %v. Skipping.", path, readErr)
					return nil
				}

				var result ResultData
				unmarshalErr := json.Unmarshal(fileData, &result)
				if unmarshalErr != nil {
					log.Printf("Error unmarshalling JSON from file %s: %v. Skipping.", path, unmarshalErr)
					return nil
				}
				result.FilePath = path
				allResults = append(allResults, result)
			}
			return nil
		})
		if err != nil {
			log.Printf("Error walking directory %s: %v", methodDir, err)
		}
	}

	log.Printf("Successfully read %d result files.", len(allResults))
	if len(allResults) == 0 {
		log.Println("No data to process. Exiting.")
		return
	}

	var Data []DataPoint
	for _, res := range allResults {
		point := DataPoint{
			FilePath:   filepath.Base(res.FilePath),
			Method:     res.RequestParams.GenerationMethod,
			Entropy:    res.ResponseMetrics.Entropy,
			FractalDim: res.ResponseMetrics.FractalDim,
		}

		var currentLandRatio float64
		for i := 3; i <= 7; i++ { // Tile types 3 (WetSand) through 7 (Forest)
			if freq, ok := res.ResponseMetrics.Frequencies[strconv.Itoa(i)]; ok {
				currentLandRatio += freq
			}
		}
		point.LandRatio = currentLandRatio

		uniquePairs := make(map[string]bool)
		if res.ResponseMetrics.Adjacency != nil {
			for typeA, innerMap := range res.ResponseMetrics.Adjacency {
				for typeB, count := range innerMap {
					if count > 0 {
						keyPart1, _ := strconv.Atoi(typeA)
						keyPart2, _ := strconv.Atoi(typeB)
						var pairKey string
						if keyPart1 <= keyPart2 {
							pairKey = fmt.Sprintf("%d-%d", keyPart1, keyPart2)
						} else {
							pairKey = fmt.Sprintf("%d-%d", keyPart2, keyPart1)
						}
						uniquePairs[pairKey] = true
					}
				}
			}
		}
		point.NumUniqueAdjacencyPairs = len(uniquePairs)

		var autocorrSumLag1 float64
		var autocorrCountLag1 int
		if val, ok := res.ResponseMetrics.Autocorr["1,0"]; ok && !math.IsNaN(val) {
			autocorrSumLag1 += val
			autocorrCountLag1++
		}
		if val, ok := res.ResponseMetrics.Autocorr["0,1"]; ok && !math.IsNaN(val) {
			autocorrSumLag1 += val
			autocorrCountLag1++
		}
		if autocorrCountLag1 > 0 {
			point.AvgAutocorrLag1 = autocorrSumLag1 / float64(autocorrCountLag1)
		} else {
			point.AvgAutocorrLag1 = math.NaN()
		}

		if len(res.ResponseMetrics.Spectrum) > 0 && len(res.ResponseMetrics.Spectrum[0]) > 0 {
			H := len(res.ResponseMetrics.Spectrum)
			W := len(res.ResponseMetrics.Spectrum[0])
			lowFreqH := H / 4
			if lowFreqH == 0 {
				lowFreqH = 1
			}
			lowFreqW := W / 4
			if lowFreqW == 0 {
				lowFreqW = 1
			}
			var lowFreqSum, totalFreqSum float64
			for r, row := range res.ResponseMetrics.Spectrum {
				for c, val := range row {
					absVal := math.Abs(val)
					if r < lowFreqH && c < lowFreqW {
						lowFreqSum += absVal
					}
					totalFreqSum += absVal
				}
			}
			if totalFreqSum > 0 {
				point.LowFreqEnergyRatio = lowFreqSum / totalFreqSum
			} else {
				point.LowFreqEnergyRatio = math.NaN()
			}
		} else {
			point.LowFreqEnergyRatio = math.NaN()
		}
		Data = append(Data, point)
	}

	log.Printf("Processed %d data points into data structure.", len(Data))
	exportDataToCSV(Data, "map_metrics_individual.csv")

	// Calculate and export averages
	calculateAndExportAverages(Data, methods, "map_metrics_averages.csv")

	log.Println("Data processing script finished.")
}

func calculateAndExportAverages(data []DataPoint, methods []string, filename string) {
	if len(data) == 0 {
		log.Println("No data to calculate averages from.")
		return
	}

	// map[methodName]map[metricName]sum
	sums := make(map[string]map[string]float64)
	// map[methodName]map[metricName]count (for non-NaN values)
	counts := make(map[string]map[string]int)

	for _, method := range methods {
		sums[method] = make(map[string]float64)
		counts[method] = make(map[string]int)
	}

	for _, point := range data {
		m := point.Method
		if _, ok := sums[m]; !ok {
			log.Printf("Warning: Method %s from data not in predefined methods list.", m)
			continue
		}

		// Summing non-NaN values and counting them
		if !math.IsNaN(point.Entropy) {
			sums[m]["Entropy"] += point.Entropy
			counts[m]["Entropy"]++
		}
		if !math.IsNaN(point.FractalDim) {
			sums[m]["FractalDim"] += point.FractalDim
			counts[m]["FractalDim"]++
		}
		if !math.IsNaN(point.LandRatio) {
			sums[m]["LandRatio"] += point.LandRatio
			counts[m]["LandRatio"]++
		}
		// NumUniqueAdjacencyPairs is int, does not have NaN
		sums[m]["NumUniqueAdjacencyPairs"] += float64(point.NumUniqueAdjacencyPairs)
		counts[m]["NumUniqueAdjacencyPairs"]++

		if !math.IsNaN(point.AvgAutocorrLag1) {
			sums[m]["AvgAutocorrLag1"] += point.AvgAutocorrLag1
			counts[m]["AvgAutocorrLag1"]++
		}
		if !math.IsNaN(point.LowFreqEnergyRatio) {
			sums[m]["LowFreqEnergyRatio"] += point.LowFreqEnergyRatio
			counts[m]["LowFreqEnergyRatio"]++
		}
	}

	var averagesData []MethodAverageData
	log.Println("\n--- Method Averages ---")
	for _, method := range methods {
		avgPoint := MethodAverageData{Method: method}
		var totalSamplesForMethod int // To find a representative sample count

		if c := counts[method]["Entropy"]; c > 0 {
			avgPoint.AvgEntropy = sums[method]["Entropy"] / float64(c)
			if c > totalSamplesForMethod {
				totalSamplesForMethod = c
			}
		} else {
			avgPoint.AvgEntropy = math.NaN()
		}

		if c := counts[method]["FractalDim"]; c > 0 {
			avgPoint.AvgFractalDim = sums[method]["FractalDim"] / float64(c)
			if c > totalSamplesForMethod {
				totalSamplesForMethod = c
			}
		} else {
			avgPoint.AvgFractalDim = math.NaN()
		}

		if c := counts[method]["LandRatio"]; c > 0 {
			avgPoint.AvgLandRatio = sums[method]["LandRatio"] / float64(c)
			if c > totalSamplesForMethod {
				totalSamplesForMethod = c
			}
		} else {
			avgPoint.AvgLandRatio = math.NaN()
		}

		if c := counts[method]["NumUniqueAdjacencyPairs"]; c > 0 {
			avgPoint.AvgNumUniqueAdjPairs = sums[method]["NumUniqueAdjacencyPairs"] / float64(c)
			if c > totalSamplesForMethod {
				totalSamplesForMethod = c
			}
		} else {
			avgPoint.AvgNumUniqueAdjPairs = math.NaN()
		}

		if c := counts[method]["AvgAutocorrLag1"]; c > 0 {
			avgPoint.AvgAvgAutocorrLag1 = sums[method]["AvgAutocorrLag1"] / float64(c)
			if c > totalSamplesForMethod {
				totalSamplesForMethod = c
			}
		} else {
			avgPoint.AvgAvgAutocorrLag1 = math.NaN()
		}

		if c := counts[method]["LowFreqEnergyRatio"]; c > 0 {
			avgPoint.AvgLowFreqEnergyRatio = sums[method]["LowFreqEnergyRatio"] / float64(c)
			if c > totalSamplesForMethod {
				totalSamplesForMethod = c
			}
		} else {
			avgPoint.AvgLowFreqEnergyRatio = math.NaN()
		}

		avgPoint.SampleCount = totalSamplesForMethod // Max count among metrics for this method
		averagesData = append(averagesData, avgPoint)

		fmt.Printf("Method: %s (Samples: %d)\n", avgPoint.Method, avgPoint.SampleCount)
		fmt.Printf("  Avg Entropy: %.4f\n", avgPoint.AvgEntropy)
		fmt.Printf("  Avg FractalDim: %.4f\n", avgPoint.AvgFractalDim)
		fmt.Printf("  Avg LandRatio: %.4f\n", avgPoint.AvgLandRatio)
		fmt.Printf("  Avg NumUniqueAdjPairs: %.2f\n", avgPoint.AvgNumUniqueAdjPairs)
		fmt.Printf("  Avg AvgAutocorrLag1: %.4f\n", avgPoint.AvgAvgAutocorrLag1)
		fmt.Printf("  Avg LowFreqEnergyRatio: %.4f\n", avgPoint.AvgLowFreqEnergyRatio)
		fmt.Println("---------------------")
	}

	exportMethodAveragesToCSV(averagesData, filename)
}

func exportDataToCSV(data []DataPoint, filename string) {
	if len(data) == 0 {
		log.Println("No individual data to export to CSV.")
		return
	}
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create CSV file %s: %v", filename, err)
	}
	defer file.Close()

	header := []string{
		"FilePath", "Method", "Entropy", "FractalDim", "LandRatio",
		"NumUniqueAdjacencyPairs", "AvgAutocorrLag1", "LowFreqEnergyRatio",
	}
	fmt.Fprintln(file, strings.Join(header, ","))

	for _, point := range data {
		row := []string{
			point.FilePath, point.Method,
			fmt.Sprintf("%.4f", point.Entropy), fmt.Sprintf("%.4f", point.FractalDim),
			fmt.Sprintf("%.4f", point.LandRatio), strconv.Itoa(point.NumUniqueAdjacencyPairs),
			fmt.Sprintf("%.4f", point.AvgAutocorrLag1), fmt.Sprintf("%.4f", point.LowFreqEnergyRatio),
		}
		fmt.Fprintln(file, strings.Join(row, ","))
	}
	log.Printf("Individual data successfully exported to %s", filename)
}

func exportMethodAveragesToCSV(data []MethodAverageData, filename string) {
	if len(data) == 0 {
		log.Println("No average data to export to CSV.")
		return
	}
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create CSV file for averages %s: %v", filename, err)
	}
	defer file.Close()

	header := []string{
		"Method", "AvgEntropy", "AvgFractalDim", "AvgLandRatio",
		"AvgNumUniqueAdjacencyPairs", "AvgAvgAutocorrLag1", "AvgLowFreqEnergyRatio", "SampleCount",
	}
	fmt.Fprintln(file, strings.Join(header, ","))

	for _, point := range data {
		row := []string{
			point.Method,
			fmt.Sprintf("%.4f", point.AvgEntropy),
			fmt.Sprintf("%.4f", point.AvgFractalDim),
			fmt.Sprintf("%.4f", point.AvgLandRatio),
			fmt.Sprintf("%.2f", point.AvgNumUniqueAdjPairs), // .2f as it's an average of counts
			fmt.Sprintf("%.4f", point.AvgAvgAutocorrLag1),
			fmt.Sprintf("%.4f", point.AvgLowFreqEnergyRatio),
			strconv.Itoa(point.SampleCount),
		}
		fmt.Fprintln(file, strings.Join(row, ","))
	}
	log.Printf("Method averages successfully exported to %s", filename)
}
