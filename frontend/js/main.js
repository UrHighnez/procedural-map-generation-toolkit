let brushSize = 3;

function initButtons() {
    document.getElementById('tools-btn').addEventListener('click', () => {
        const toolsPanel = document.getElementById('tools-panel');
        toolsPanel.classList.toggle('tools-panel-visible');
        toolsPanel.classList.toggle('hidden');
    });

    document.getElementById('save-btn').addEventListener('click', saveCanvas);
    document.getElementById('load-btn').addEventListener('click', loadCanvas);

    document.getElementById('generate-btn').addEventListener('click', generateCanvas);

    document.getElementById('paint-water').addEventListener('click', () => setPaintColor('#00507f'));
    document.getElementById('paint-land').addEventListener('click', () => setPaintColor('#ffd675'));
    document.getElementById('paint-forest').addEventListener('click', () => setPaintColor('#2c7519'));

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

    randomnessValue.innerText = randomnessFactor.toFixed(2);

    randomnessSlider.addEventListener('input', function () {
        randomnessFactor = parseFloat(randomnessSlider.value);
        randomnessValue.innerText = randomnessFactor.toFixed(2);
    });

    window.getRandomnessFactor = function () {
        return randomnessFactor;
    }
}

document.addEventListener('DOMContentLoaded', () => {
    initGrid();
    initButtons();
    initPainting();
});

async function saveCanvas() {
    const canvas = document.getElementById('paint-canvas');
    const imageData = canvas.toDataURL('image/png');

    // Send the image data to the backend
    try {
        const response = await fetch('/save', {
            generationMethod: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({imageData}),
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
    const width = canvas.width; // Set the desired width
    const height = canvas.height; // Set the desired height

    let grid = [];

    const rows = Math.ceil(height / TileSize);
    const cols = Math.ceil(width / TileSize);

    const paintedTiles = getPaintedTiles();

    let iterations = Number(document.getElementById('iteration-slider').value);
    const randomness = window.getRandomnessFactor();

    const generationMethod = document.getElementById('generation-method').value;

    try {
        const response = await fetch('/generate', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({
                width: cols,
                height: rows,
                prevGrid: grid,
                paintedTiles: paintedTiles,
                randomnessFactor: randomness,
                iterations,
                generationMethod: generationMethod
            }),
        });


        if (response.ok) {
            const mapData = await response.json();
            grid = mapData;
            console.log('Server response: ', mapData);


            // Render the generated map on the canvas
            const canvas = document.getElementById('paint-canvas');
            const ctx = canvas.getContext('2d');

            let isObjectFormat = !!(mapData[0] && typeof mapData[0][0] === "object" && "Color" in mapData[0][0]);

            for (let y = 0; y < mapData.length; y++) {
                for (let x = 0; x < mapData[y].length; x++) {

                    let colorCode;
                    if (isObjectFormat) {
                        colorCode = mapData[y][x].Color;
                    } else {
                        colorCode = mapData[y][x];
                    }

                    switch (colorCode) {
                        case 0:
                            ctx.fillStyle = '#00507f';
                            break;
                        case 1:
                            ctx.fillStyle = '#1085bc';
                            break;
                        case 2:
                            ctx.fillStyle = '#3eb3e6';
                            break;
                        case 3:
                            ctx.fillStyle = '#b59752';
                            break;
                        case 4:
                            ctx.fillStyle = '#ffd675';
                            break;
                        case 5:
                            ctx.fillStyle = '#78e85b';
                            break;
                        case 6:
                            ctx.fillStyle = '#4caf32';
                            break;
                        case 7:
                            ctx.fillStyle = '#2c7519';
                            break;
                        default:
                            ctx.fillStyle = '#000000';
                            break;
                    }
                    ctx.fillRect(x * TileSize, y * TileSize, TileSize, TileSize);
                }
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
        } else if (r === 62 && g === 191 && b === 230) {
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
                // console.log(`Detected color ${color} at (${x}, ${y})`);
                tiles[y][x] = color;
            }
        }
    }

    return tiles;
}
