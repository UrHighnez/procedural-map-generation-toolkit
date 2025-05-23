# Procedural Map Generation Toolkit

## Description

This project provides a modular toolkit for the procedural generation of 2D tile-based maps. It supports multiple
algorithms and automatically computes metrics to evaluate the generated worlds.

### Supported Algorithms

* **Multi-Layered Cellular Automata (MLCA)**
* **Discrete Perlin Noise**
* **Wave Function Collapse (WFC)**

### Computed Metrics

* **Tile Entropy** (information content)
* **Adjacency Matrix** (neighbor distributions)
* **Cluster Sizes** (connected regions)
* **Tile Type Frequencies**
* **Fractal Dimension** *(planned)*
* **Spectral Analysis** *(planned)*
* **Pattern Repetition** *(planned)*
* **Gradient Distribution** *(planned)*

### Features

* **Interactive Painting:** Paint individual tiles in the browser canvas
* **Map Save/Load:** Export and import maps as PNG files
* **Real-time Metrics Display:** Entropy, cluster sizes, adjacency, and frequencies

## Architecture

* **Backend (Go):** Echo web server with endpoints:

    * `/generate` generates maps via selected algorithm
    * `/save` saves canvas as PNG
    * `/load` lists and loads saved maps
      Modules: `ca`, `mlca`, `noise`, `wfc`, `metrics`
* **Frontend (JavaScript/HTML/CSS):**

    * HTML5 Canvas for rendering
    * UI controls for algorithm selection, parameters, painting tools
    * Display panel for computed metrics

## Installation

```bash
# Clone repository
git clone https://github.com/UrHighnez/procedural-map-generation-toolkit.git
cd procedural-map-generation-toolkit

# Download dependencies
go mod download
```

## Usage

1. **Start server**

   ```bash
   go run .\backend\main.go
   ```

   The server listens on port 8000 by default.

2. **Open browser**
   Navigate to [http://localhost:8000](http://localhost:8000) to access the UI.

3. **Configure parameters**

    * Choose generation method
    * Set iterations, randomness, and noise parameters

4. **Generate map**
   Click “Generate” to render a new map on the canvas.

5. **Paint tiles**
   Use the brush tools to customize tile colors.

6. **View metrics**
   The panel below the canvas shows entropy, cluster sizes, adjacency, and tile frequencies.

7. **Save map**
   Click “Save” to export the current canvas as a PNG in the `saved_maps` folder.

8. **Load map**
   Click “Load” to choose and import a previously saved map.

## Development & Testing

* **Format code**: `go fmt ./...`
* **Lint**: `golangci-lint run`
* **Tests**: (unit tests for metrics modules pending)

## Contributing

1. Fork the repository.
2. Create a feature branch: `git checkout -b feature/your-feature`
3. Implement your changes and add tests.
4. Commit and push to your fork.
5. Open a Pull Request describing your changes.

## License

This project is released under the MIT License. See [LICENSE](LICENSE) for details.
