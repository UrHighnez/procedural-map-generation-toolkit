import {getClusters} from './utils.js';

// --- Metrics Visualization Functions ---

function visualizeEntropy(data) {
    document.getElementById('entropy').textContent = data.entropy.toFixed(2);
}

function visualizeFractalDimension(data) {
    document.getElementById('fractalDim').textContent = data.fractalDim.toFixed(2);
}

function visualizeFrequencies(data) {
    const container = document.getElementById('frequencies');
    container.innerHTML = '';
    // Use CSS Grid for a clean, aligned layout of labels and bars.
    container.style.display = 'grid';
    container.style.gridTemplateColumns = 'auto 1fr'; // Col 1 for labels, Col 2 for bars
    container.style.gap = '4px 8px'; // Add some space between rows and columns
    container.style.alignItems = 'center';

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
        container.appendChild(label);
        container.appendChild(barContainer);
    });
}

function visualizeClusters(data) {
    const container = document.getElementById('clusters');
    container.innerHTML = '';
    const clusters = getClusters(data.grid);

    // Step 1: Process the raw cluster list into a summary object.
    // Group by tileType and calculate count, totalSize, and maxSize.
    const summarizedClusters = {};
    clusters.forEach(({tileType, size}) => {
        if (!summarizedClusters[tileType]) {
            summarizedClusters[tileType] = {
                count: 0, totalSize: 0, maxSize: 0,
            };
        }
        summarizedClusters[tileType].count++;
        summarizedClusters[tileType].totalSize += size;
        if (size > summarizedClusters[tileType].maxSize) {
            summarizedClusters[tileType].maxSize = size;
        }
    });

    // Step 2: Build the HTML display from the summarized data.
    // Use CSS Grid again for a clean, multi-column layout.
    container.style.display = 'grid';
    container.style.gridTemplateColumns = 'auto auto 1fr'; // Icon | Count | Details
    container.style.gap = '4px 8px';
    container.style.alignItems = 'center';

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
        container.appendChild(icon);
        container.appendChild(countSpan);
        container.appendChild(detailsSpan);
    });
}

function visualizeAdjacency(data) {
    const container = document.getElementById('adjacency');
    container.innerHTML = '';
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

        // The first cell in the row is the tile icon header.
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

    container.appendChild(table);
}

function visualizeAutocorrelation(data) {
    const container = document.getElementById('autocorr');
    container.innerHTML = '';
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
        container.appendChild(table);
    }
}

function visualizeSpectrum(data) {
    const container = document.getElementById('spectrum');
    container.innerHTML = '';
    // Remove default line-height to make the grid compact
    container.style.lineHeight = '0';

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
        container.appendChild(table);
    }
}

/**
 * Updates the entire metrics panel with new data from the server.
 * @param {object} data - The response object from the /generate endpoint.
 */
export function updateMetricsPanel(data) {
    if (!data) return;
    visualizeEntropy(data);
    visualizeFractalDimension(data);
    visualizeFrequencies(data);
    visualizeClusters(data);
    visualizeAdjacency(data);
    visualizeAutocorrelation(data);
    visualizeSpectrum(data);
}

/**
 * Initializes all buttons and sliders in the UI.
 * @param {object} callbacks - An object with callback functions for button clicks.
 */
export function initControls(callbacks) {
    // --- Toolbar Buttons ---
    document.getElementById('tools-btn').addEventListener('click', () => {
        document.getElementById('tools-panel').classList.toggle('hidden');
    });
    document.getElementById('generate-btn').addEventListener('click', callbacks.onGenerate);
    document.getElementById('save-btn').addEventListener('click', callbacks.onSave);
    document.getElementById('load-btn').addEventListener('click', callbacks.onLoad);

    // --- Paint Color Buttons ---
    const paintConfig = [{id: 'paint-water', tileIndex: 0}, {id: 'paint-sand', tileIndex: 4}, {
        id: 'paint-forest', tileIndex: 7
    },];
    paintConfig.forEach(({id, tileIndex}) => {
        document.getElementById(id).addEventListener('click', () => {
            const color = window.tileColors[tileIndex];
            window.setPaintColor(color);
        });
    });

    // --- Sliders ---
    const brushSizeSlider = document.getElementById('brushSize-slider');
    const brushSizeValue = document.getElementById('brushSize-value');
    brushSizeSlider.addEventListener('input', (event) => {
        window.brushSize = parseInt(event.target.value);
        brushSizeValue.textContent = window.brushSize;
    });
    brushSizeValue.textContent = brushSizeSlider.value;
    window.brushSize = parseInt(brushSizeSlider.value);

    // Setup for Iteration Slider
    const iterationSlider = document.getElementById('iteration-slider');
    const iterationValue = document.getElementById('iteration-value');
    iterationSlider.addEventListener('input', () => {
        iterationValue.textContent = iterationSlider.value;
    });
    iterationValue.textContent = iterationSlider.value; // Set initial display value

    // Setup for Randomness Slider
    const randomnessSlider = document.getElementById('randomness-slider');
    const randomnessValue = document.getElementById('randomness-value');
    randomnessSlider.addEventListener('input', () => {
        // toFixed(1) ensures it displays as 0.0, 0.1, etc.
        randomnessValue.textContent = parseFloat(randomnessSlider.value).toFixed(1);
    });
    randomnessValue.textContent = parseFloat(randomnessSlider.value).toFixed(1); // Set initial display value
}