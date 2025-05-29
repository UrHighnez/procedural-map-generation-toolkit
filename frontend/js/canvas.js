import {TileSize} from './grid.js';

/**
 * Renders the generated grid onto the paint canvas.
 * @param {HTMLCanvasElement} canvas - The canvas element to draw on.
 * @param {number[][]} grid - The 2D array of tile indices.
 * @param {string[]} tileColors - The array of color strings.
 */
export function renderGrid(canvas, grid, tileColors) {
    const ctx = canvas.getContext('2d');
    grid.forEach((row, y) => {
        row.forEach((colorCode, x) => {
            ctx.fillStyle = tileColors[colorCode] || '#000000';
            ctx.fillRect(x * TileSize, y * TileSize, TileSize, TileSize);
        });
    });
}

/**
 * Reads the painted tiles from the canvas and converts them to a grid of tile indices.
 * @param {HTMLCanvasElement} canvas - The canvas element to read from.
 * @returns {number[][]} A 2D array of tile indices.
 */
export function getPaintedTiles(canvas) {
    const ctx = canvas.getContext('2d');
    const rows = Math.ceil(canvas.height / TileSize);
    const cols = Math.ceil(canvas.width / TileSize);
    const tiles = Array.from({length: rows}, () => Array(cols).fill(-1));

    for (let y = 0; y < rows; y++) {
        for (let x = 0; x < cols; x++) {
            const imgData = ctx.getImageData(x * TileSize, y * TileSize, 1, 1);
            tiles[y][x] = getColorIndex(imgData);
        }
    }
    return tiles;
}

/**
 * Determines the tile index for a given pixel's color data.
 * @param {ImageData} imageData - The ImageData from a canvas pixel.
 * @returns {number} The corresponding tile index, or -1 if not found.
 */
function getColorIndex(imageData) {
    const [r, g, b] = imageData.data;
    const hexColor = "#" + ((1 << 24) + (r << 16) + (g << 8) + b).toString(16).slice(1);

    const index = window.tileColors.findIndex(color => color === hexColor);
    return index;
}