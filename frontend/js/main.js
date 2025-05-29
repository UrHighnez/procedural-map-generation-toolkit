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
            document.getElementById('fractalDim').textContent = data.fractalDim.toFixed(2);

            // CLUSTER
            const clustersContainer = document.getElementById('clusters');
            clustersContainer.innerHTML = ''; // Clear previous content

            const rawGrid = Array.isArray(data) ? data : data.grid;
            // The getClusters function is already in your main.js file
            const clusters = getClusters(rawGrid);

            // Step 1: Process the raw cluster list into a summary object.
            // We'll group by tileType and calculate count, totalSize, and maxSize.
            const summarizedClusters = {};
            clusters.forEach(({tileType, size}) => {
                if (!summarizedClusters[tileType]) {
                    summarizedClusters[tileType] = {
                        count: 0,
                        totalSize: 0,
                        maxSize: 0,
                    };
                }
                summarizedClusters[tileType].count++;
                summarizedClusters[tileType].totalSize += size;
                if (size > summarizedClusters[tileType].maxSize) {
                    summarizedClusters[tileType].maxSize = size;
                }
            });

            // Step 2: Build the HTML display from the summarized data.
            // We'll use CSS Grid again for a clean, multi-column layout.
            clustersContainer.style.display = 'grid';
            clustersContainer.style.gridTemplateColumns = 'auto auto 1fr'; // Icon | Count | Details
            clustersContainer.style.gap = '4px 8px';
            clustersContainer.style.alignItems = 'center';

            // Sort by tile index for a consistent display order.
            const sortedClusterEntries = Object.entries(summarizedClusters).sort(([a], [b]) => +a - +b);

            sortedClusterEntries.forEach(([tileType, summary]) => {
                const idx = +tileType;
                const color = data.colors[idx] || '#cccccc';
                const avgSize = (summary.totalSize / summary.count).toFixed(1);

                // Column 1: The color swatch for the tile type
                const icon = document.createElement('div');
                icon.innerHTML = `
                    <span style="
                        display: inline-block;
                        width: 0.8em;
                        height: 0.8em;
                        vertical-align: middle;
                        background-color: ${color};
                        border: 1px solid #333;
                    "></span>
                `;

                // Column 2: The number of clusters found (e.g., "3x")
                const countSpan = document.createElement('span');
                countSpan.style.textAlign = 'right';
                countSpan.style.fontWeight = 'bold';
                countSpan.textContent = `${summary.count}x`;

                // Column 3: Details about the max and average size
                const detailsSpan = document.createElement('span');
                detailsSpan.style.fontSize = '0.9em';
                detailsSpan.style.color = '#444';
                detailsSpan.textContent = `(Max: ${summary.maxSize}, Avg: ${avgSize})`;

                // Add all columns to the grid container
                clustersContainer.appendChild(icon);
                clustersContainer.appendChild(countSpan);
                clustersContainer.appendChild(detailsSpan);
            });


            // FREQUENCY
            const frequenciesContainer = document.getElementById('frequencies');
            frequenciesContainer.innerHTML = ''; // Clear previous content

            // Use CSS Grid for a clean, aligned layout of labels and bars.
            frequenciesContainer.style.display = 'grid';
            frequenciesContainer.style.gridTemplateColumns = 'auto 1fr'; // Col 1 for labels, Col 2 for bars
            frequenciesContainer.style.gap = '4px 8px'; // Add some space between rows and columns
            frequenciesContainer.style.alignItems = 'center';

            // Sort by tile index to ensure the order is always the same.
            const sortedFrequencies = Object.entries(data.frequencies).sort(([a], [b]) => +a - +b);

            sortedFrequencies.forEach(([tileIdx, p]) => {
                // Don't display a bar for tile types with zero frequency.
                if (p <= 0) {
                    return;
                }

                const idx = +tileIdx;
                const color = data.colors[idx] || '#cccccc';
                const percent = (p * 100).toFixed(1);

                // --- Create the Label (Color Swatch + Percentage Text) ---
                const label = document.createElement('div');
                label.style.whiteSpace = 'nowrap'; // Prevent text from wrapping
                label.style.textAlign = 'right';
                label.innerHTML = `
                    <span style="
                        display: inline-block;
                        width: 0.8em;
                        height: 0.8em;
                        vertical-align: middle;
                        background-color: ${color};
                        border: 1px solid #333;
                    "></span>
                    <span style="font-size: 0.9em; vertical-align: middle;">
                        ${percent}%
                    </span>
                `;

                // --- Create the Bar ---
                // 1. A container div that acts as the background "track" for the bar.
                const barContainer = document.createElement('div');
                barContainer.style.width = '100%';
                barContainer.style.backgroundColor = '#e0e0e0';
                barContainer.style.borderRadius = '3px';
                barContainer.style.height = '1.2em';

                // 2. The actual colored bar, with its width set by the frequency.
                const bar = document.createElement('div');
                bar.style.width = `${p * 100}%`;
                bar.style.height = '100%';
                bar.style.backgroundColor = color;
                bar.style.borderRadius = '3px';
                // Add a nice transition for when the values change.
                bar.style.transition = 'width 0.4s ease-in-out';

                barContainer.appendChild(bar);

                // Add the new elements to the grid container
                frequenciesContainer.appendChild(label);
                frequenciesContainer.appendChild(barContainer);
            });


            // ADJACENCY
            const adjacencyContainer = document.getElementById('adjacency');
            adjacencyContainer.innerHTML = ''; // Clear previous content

            const adjacencyData = data.adjacency;
            const colors = data.colors;
            const numTypes = colors.length;

            // Step 1: Find the maximum count to normalize for the heatmap color.
            let maxCount = 0;
            for (let i = 0; i < numTypes; i++) {
                for (let j = 0; j < numTypes; j++) {
                    const count = (adjacencyData[i] || [])[j] || 0;
                    if (count > maxCount) {
                        maxCount = count;
                    }
                }
            }

            // Step 2: Create the table structure.
            const table = document.createElement('table');
            table.style.borderCollapse = 'collapse';
            table.style.fontSize = '0.9em';
            table.style.margin = 'auto';

            // Step 3: Create the top header row with tile icons as column labels.
            const headerRow = document.createElement('tr');
            headerRow.appendChild(document.createElement('th')); // Empty top-left corner
            for (let j = 0; j < numTypes; j++) {
                const th = document.createElement('th');
                th.style.padding = '2px';
                th.innerHTML = `<div title="${j}" style="width: 1.2em; height: 1.2em; background-color: ${colors[j]}; border: 1px solid #555;"></div>`;
                headerRow.appendChild(th);
            }
            table.appendChild(headerRow);

            // Step 4: Create a data row for each tile type.
            for (let i = 0; i < numTypes; i++) {
                const row = document.createElement('tr');

                // First cell in the row is the tile icon header.
                const th = document.createElement('th');
                th.style.padding = '2px';
                th.innerHTML = `<div title="${i}" style="width: 1.2em; height: 1.2em; background-color: ${colors[i]}; border: 1px solid #555;"></div>`;
                row.appendChild(th);

                // Create the data cells for the row.
                for (let j = 0; j < numTypes; j++) {
                    const td = document.createElement('td');
                    const count = (adjacencyData[i] || [])[j] || 0;

                    td.textContent = count;
                    td.style.padding = '5px';
                    td.style.textAlign = 'center';
                    td.style.border = '1px solid #ddd';
                    td.title = `${count} pairs of type ${i} adjacent to type ${j}`;

                    // Apply heatmap color. Using sqrt helps visualize variance better.
                    if (maxCount > 0) {
                        const intensity = Math.sqrt(count / maxCount);
                        // Using a blue color scale: transparent for 0, dark blue for maxCount.
                        td.style.backgroundColor = `rgba(0, 80, 200, ${intensity})`;
                        // Make text white on darker backgrounds for better readability.
                        if (intensity > 0.6) {
                            td.style.color = 'white';
                        }
                    }
                    row.appendChild(td);
                }
                table.appendChild(row);
            }

            adjacencyContainer.appendChild(table);

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