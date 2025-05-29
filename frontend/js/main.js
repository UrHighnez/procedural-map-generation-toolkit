async function loadColors() {
    const res = await fetch('/colors');
    if (!res.ok) throw new Error('Colors load failed');
    window.tileColors = await res.json();
}

document.addEventListener('DOMContentLoaded', async () => {
    await loadColors();
    initGrid();
    initButtons();
    initPainting();
});

let brushSize = 3;

const paintConfig = [{id: 'paint-water', tileIndex: 0}, {id: 'paint-sand', tileIndex: 4}, {
    id: 'paint-forest', tileIndex: 7
},];

function initButtons() {
    document.getElementById('tools-btn').addEventListener('click', () => {
        const toolsPanel = document.getElementById('tools-panel');
        toolsPanel.classList.toggle('tools-panel-visible');
        toolsPanel.classList.toggle('hidden');
    });

    document.getElementById('save-btn').addEventListener('click', saveCanvas);
    document.getElementById('load-btn').addEventListener('click', loadCanvas);

    document.getElementById('generate-btn').addEventListener('click', generateCanvas);

    paintConfig.forEach(({id, tileIndex}) => {
        const btn = document.getElementById(id);
        if (!btn) return;
        btn.addEventListener('click', () => {
            const color = window.tileColors[tileIndex];
            setPaintColor(color);
        });
    });

    const brushSizeSlider = document.getElementById('brushSize-slider');
    const brushSizeValue = document.getElementById('brushSize-value');

    brushSizeSlider.addEventListener('input', function (event) {
        brushSize = parseInt(event.target.value);
        brushSizeValue.textContent = brushSize;
    });


    brushSizeValue.textContent = brushSizeValue.value = brushSize;

    const iterationSlider = document.getElementById('iteration-slider');
    const iterationValue = document.getElementById('iteration-value');

    iterationValue.textContent = iterationSlider.value;

    iterationSlider.addEventListener('input', function () {
        iterationValue.textContent = iterationSlider.value;
    });

    const randomnessSlider = document.getElementById('randomness-slider');
    const randomnessValue = document.getElementById('randomness-value');
    let randomnessFactor = parseFloat(randomnessSlider.value);

    randomnessValue.innerText = randomnessFactor.toFixed(1);

    randomnessSlider.addEventListener('input', function () {
        randomnessFactor = parseFloat(randomnessSlider.value);
        randomnessValue.innerText = randomnessFactor.toFixed(1);
    });

    window.getRandomnessFactor = function () {
        return randomnessFactor;
    }
}


async function saveCanvas() {
    const canvas = document.getElementById('paint-canvas');
    const imageData = canvas.toDataURL('image/png');

    // Send the image data to the backend
    try {
        const response = await fetch('/save', {
            method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({imageData}),
        });

        if (response.ok) {
            console.log('Canvas image saved successfully.');
        } else {
            console.error('Failed to save the canvas image.');
        }
    } catch (error) {
        console.error('Error while saving the canvas image:', error);
    }
}

async function loadCanvas() {
    try {
        const response = await fetch('/load');
        if (response.ok) {
            const imageFiles = await response.json();
            // Prompt the user to choose an image file to load
            const selectedFile = prompt('Choose a map file to load:', imageFiles.join(', '));
            if (selectedFile) {
                const img = new Image();
                img.onload = () => {
                    const canvas = document.getElementById('paint-canvas');
                    const ctx = canvas.getContext('2d');

                    ctx.clearRect(0, 0, canvas.width, canvas.height); // Clear the background canvas
                    ctx.drawImage(img, 0, 0);
                };
                img.src = `/maps/${selectedFile}`;
            }
        } else {
            console.error('Failed to load map images.');
        }
    } catch (error) {
        console.error('Error while loading map images:', error);
    }
}

async function generateCanvas() {
    const canvas = document.getElementById('paint-canvas');
    const width = canvas.width;
    const height = canvas.height;

    let grid = [];

    const rows = Math.ceil(height / TileSize);
    const cols = Math.ceil(width / TileSize);

    const paintedTiles = getPaintedTiles();

    let iterations = Number(document.getElementById('iteration-slider').value);
    const randomness = window.getRandomnessFactor();

    const generationMethod = document.getElementById('generation-method').value;

    try {
        const response = await fetch('/generate', {
            method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({
                width: cols, height: rows, prevGrid: grid, paintedTiles: paintedTiles, randomnessFactor: randomness,

                noiseScale: 0.5, noiseOctaves: 4, noisePersistence: 0.9, noiseLacunarity: 1.8,

                iterations,

                generationMethod: generationMethod
            }),
        });


        if (response.ok) {
            const data = await response.json();
            console.log('Server response: ', data);

            const grid = Array.isArray(data) ? data : data.grid;

            // Render the generated map on the canvas
            const canvas = document.getElementById('paint-canvas');
            const ctx = canvas.getContext('2d');
            const tileColors = data.colors;

            for (let y = 0; y < grid.length; y++) {
                for (let x = 0; x < grid[y].length; x++) {

                    const colorCode = grid[y][x];

                    // Fallback
                    ctx.fillStyle = tileColors[colorCode] || '#000000';

                    ctx.fillRect(x * TileSize, y * TileSize, TileSize, TileSize);
                }
            }

            // ENTROPY
            document.getElementById('entropy').textContent = data.entropy.toFixed(2);

            // FRACTAL DIMENSION
            // data.fractalDim is a number
            document.getElementById('fractalDim').textContent = data.fractalDim.toFixed(2);

            // CLUSTER
            const rawGrid = Array.isArray(data) ? data : data.grid;
            const clusters = getClusters(rawGrid);
            const clusterStrings = clusters.map(({tileType, size}) => {
                const color = data.colors[tileType] || '#cccccc';
                return `
                    <span style="
                        display:inline-block;
                        width:0.5em; height:0.5em;
                        vertical-align:middle;
                        background-color:${color};
                    "></span>
                    ${size}
                `;
            });
            document.getElementById('clusters').innerHTML = clusterStrings.join(' | ');


            // FREQUENCY
            const freqStrings = Object.entries(data.frequencies).map(([tileIdx, p]) => {
                const idx = +tileIdx;
                const color = data.colors[idx] || '#cccccc';
                const percent = (p * 100).toFixed(1);
                return `
                    <span style="
                        display:inline-block;
                        width:0.5em; height:0.5em;
                        vertical-align:middle;
                        background-color:${color};
                    "></span>
                    ${idx}: ${percent}%
                `;
            });
            document.getElementById('frequencies').innerHTML = freqStrings.join(' | ');


            // ADJACENCY
            const adjStrings = [];
            const C = data.colors.length;
            for (let i = 0; i < C; i++) {
                for (let j = 0; j < C; j++) {
                    const count = (data.adjacency[i] || [])[j] || 0;
                    if (!count) continue;
                    const ci = data.colors[i] || '#cccccc';
                    const cj = data.colors[j] || '#cccccc';
                    adjStrings.push(`
                        <span style="
                            display:inline-block;
                            width:0.5em; height:0.5em;
                            background-color:${ci};
                        "></span>
                        →
                        <span style="
                            display:inline-block;
                            width:0.5em; height:0.5em;
                            background-color:${cj};
                        "></span>
                        ${count}
                    `);
                }
            }
            document.getElementById('adjacency').innerHTML = adjStrings.join(' | ');

            // AUTOCORRELATION
            // data.autocorr is an object { "dx,dy": value, ... }
            const autocorrContainer = document.getElementById('autocorr');
            autocorrContainer.innerHTML = ''; // Clear previous content

            const lags = Object.keys(data.autocorr);

            if (lags.length > 0) {
                const lagCoords = lags.map(k => k.split(',').map(Number));
                const all_dx = lagCoords.map(l => l[0]);
                const all_dy = lagCoords.map(l => l[1]);
                const min_dx = Math.min(...all_dx);
                const max_dx = Math.max(...all_dx);
                const min_dy = Math.min(...all_dy);
                const max_dy = Math.max(...all_dy);

                const table = document.createElement('table');
                table.style.borderCollapse = 'collapse';
                table.style.margin = 'auto'; // Center the table visually

                for (let dy = min_dy; dy <= max_dy; dy++) {
                    const row = document.createElement('tr');
                    for (let dx = min_dx; dx <= max_dx; dx++) {
                        const cell = document.createElement('td');
                        const key = `${dx},${dy}`;
                        const value = data.autocorr[key];

                        // Basic cell styling
                        cell.style.width = '1em';
                        cell.style.height = '1em';
                        cell.style.border = '1px solid #555';

                        if (value !== undefined) {
                            // Map the correlation value (typically [0, 1]) to a grayscale color.
                            // A value of 1.0 will be white, 0.0 will be black.
                            const intensity = Math.max(0, Math.min(1, value)); // Clamp value to [0, 1]
                            const colorVal = Math.round(intensity * 255);
                            cell.style.backgroundColor = `rgb(${colorVal}, ${colorVal}, ${colorVal})`;

                            // Add a tooltip to show the precise value on hover
                            cell.title = `Lag (${dx}, ${dy}): ${value.toFixed(3)}`;
                        } else {
                            // Style for lags that are not in the data, if any
                            cell.style.backgroundColor = '#222';
                        }

                        // Highlight the center cell (lag 0,0) which is always 1.0
                        if (dx === 0 && dy === 0) {
                            cell.style.outline = '1px solid red';
                            cell.style.outlineOffset = '-1px';
                        }

                        row.appendChild(cell);
                    }
                    table.appendChild(row);
                }
                autocorrContainer.appendChild(table);
            }

            // SPECTRAL ANALYSIS
            // data.spectrum is a 2D array [[...], [...], ...]
            const spectrumContainer = document.getElementById('spectrum');
            spectrumContainer.innerHTML = ''; // Clear previous content
            // Remove default line-height to make the grid compact
            spectrumContainer.style.lineHeight = '0';

            const spectrum = data.spectrum;

            if (spectrum && spectrum.length > 0 && spectrum[0].length > 0) {
                // Spectrums often have a huge dynamic range. A log scale is best for visualization.
                // First, find the maximum log-transformed value for normalization.
                let maxLogVal = 0;
                for (let y = 0; y < spectrum.length; y++) {
                    for (let x = 0; x < spectrum[y].length; x++) {
                        // Use log(1 + value) to avoid issues with log(0)
                        const logVal = Math.log(1 + spectrum[y][x]);
                        if (logVal > maxLogVal) {
                            maxLogVal = logVal;
                        }
                    }
                }

                const table = document.createElement('table');
                table.style.borderCollapse = 'collapse';
                table.style.margin = 'auto';

                // Build the table row by row
                for (let y = 0; y < spectrum.length; y++) {
                    const row = document.createElement('tr');
                    for (let x = 0; x < spectrum[y].length; x++) {
                        const cell = document.createElement('td');
                        const value = spectrum[y][x];

                        // Normalize the log-transformed value to get an intensity from 0 to 1
                        const logVal = Math.log(1 + value);
                        const intensity = maxLogVal > 0 ? (logVal / maxLogVal) : 0;

                        // Map intensity to a grayscale color (0=black, 1=white)
                        const colorVal = Math.round(intensity * 255);
                        cell.style.backgroundColor = `rgb(${colorVal}, ${colorVal}, ${colorVal})`;

                        // Style cells to be small, square, and without padding for a clean look
                        const cellSize = Math.max(1, Math.floor(128 / Math.max(spectrum.length, spectrum[0].length)));
                        cell.style.width = `${cellSize}px`;
                        cell.style.height = `${cellSize}px`;
                        cell.style.padding = '0';

                        // Add a tooltip to show the original value in scientific notation
                        cell.title = `Freq (${x}, ${y}): ${value.toExponential(2)}`;

                        row.appendChild(cell);
                    }
                    table.appendChild(row);
                }
                spectrumContainer.appendChild(table);
            }

        }

        if (!response.ok) {
            const text = await response.text();
            console.error('Failed to generate the map: ', response.status, text);
        }
    } catch (error) {
        console.error('Error while generating the map: ', error);
    }
}

function getColor(imageData) {
    for (let i = 0; i < imageData.data.length; i += 4) {
        const r = imageData.data[i];
        const g = imageData.data[i + 1];
        const b = imageData.data[i + 2];

        if (r === 0 && g === 80 && b === 127) {
            return 0;
        } else if (r === 16 && g === 133 && b === 188) {
            return 1;
        } else if (r === 62 && g === 179 && b === 230) {
            return 2;
        } else if (r === 181 && g === 151 && b === 82) {
            return 3;
        } else if (r === 255 && g === 214 && b === 117) {
            return 4;
        } else if (r === 120 && g === 232 && b === 91) {
            return 5;
        } else if (r === 76 && g === 175 && b === 50) {
            return 6;
        } else if (r === 44 && g === 117 && b === 25) {
            return 7;
        }
    }

    return -1; // Empty tile
}

function getPaintedTiles() {
    const canvas = document.getElementById('paint-canvas');
    const ctx = canvas.getContext('2d');
    const width = canvas.width;
    const height = canvas.height;

    const rows = Math.ceil(height / TileSize);
    const cols = Math.ceil(width / TileSize);

    const tiles = new Array(rows).fill(-1).map(() => new Array(cols).fill(-1));

    for (let y = 0; y < rows; y++) {
        for (let x = 0; x < cols; x++) {
            const imgData = ctx.getImageData(x * TileSize, y * TileSize, TileSize, TileSize);
            const color = getColor(imgData);

            if (color !== -1) {
                tiles[y][x] = color;
            }
        }
    }

    return tiles;
}

function getClusters(grid) {
    const rows = grid.length, cols = grid[0].length;
    const seen = Array.from({length: rows}, () => Array(cols).fill(false));
    const clusters = [];

    for (let y = 0; y < rows; y++) {
        for (let x = 0; x < cols; x++) {
            if (seen[y][x]) continue;
            const type = grid[y][x];
            let size = 0;
            const stack = [[x, y]];
            seen[y][x] = true;

            while (stack.length) {
                const [cx, cy] = stack.pop();
                size++;
                [[1, 0], [-1, 0], [0, 1], [0, -1]].forEach(([dx, dy]) => {
                    const nx = cx + dx, ny = cy + dy;
                    if (nx >= 0 && nx < cols && ny >= 0 && ny < rows && !seen[ny][nx] && grid[ny][nx] === type) {
                        seen[ny][nx] = true;
                        stack.push([nx, ny]);
                    }
                });
            }

            clusters.push({tileType: type, size});
        }
    }

    return clusters;
}