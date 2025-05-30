import * as api from './api.js';
import * as ui from './ui.js';
import {updateMetricsPanel} from './ui.js';
import {getPaintedTiles, renderGrid} from './canvas.js';
import {initGrid, TileSize} from './grid.js';
import {initExportButtons} from './export.js';

// --- Main Application State ---
const state = {
    paintCanvas: null,
};

// --- Event Handlers ---

async function handleGenerate() {
    try {
        const paintCanvas = document.getElementById('paint-canvas');

        const params = {
            width: Math.ceil(paintCanvas.width / TileSize),
            height: Math.ceil(paintCanvas.height / TileSize),
            paintedTiles: getPaintedTiles(paintCanvas),
            generationMethod: document.getElementById('generation-method').value,

            // Read slider values
            iterations: Number(document.getElementById('iteration-slider').value),
            randomnessFactor: parseFloat(document.getElementById('randomness-slider').value),

            // Example noise params
            noiseScale: 0.5,
            noiseOctaves: 4,
            noisePersistence: 0.9,
            noiseLacunarity: 1.8,
        };

        const data = await api.generate(params);
        console.log('Server response: ', data);

        renderGrid(paintCanvas, data.grid, data.colors);
        updateMetricsPanel(data);

    } catch (error) {
        console.error('Error in handleGenerate: ', error);
        alert(error.message);
    }
}

async function handleSave() {
    try {
        await api.saveCanvas(state.paintCanvas);
    } catch (error) {
        console.error('Error in handleSave: ', error);
        alert(error.message);
    }
}

async function handleLoad() {
    try {
        await api.loadCanvasTo(state.paintCanvas);
    } catch (error) {
        console.error('Error in handleLoad: ', error);
        alert(error.message);
    }
}

// --- Initialization ---

document.addEventListener('DOMContentLoaded', async () => {
    try {
        // Load critical data
        window.tileColors = await api.getColors();

        // Initialize UI components
        initGrid();
        state.paintCanvas = document.getElementById('paint-canvas');
        // initPainting() is already called from its own script

        // Connect UI controls to handlers
        ui.initControls({
            onGenerate: handleGenerate, onSave: handleSave, onLoad: handleLoad,
        });

        initExportButtons();

    } catch (error) {
        console.error("Initialization failed:", error);
        alert("Could not initialize the application. Please check the console and refresh.");
    }
});