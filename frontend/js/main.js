let brushSize = 3;

function initButtons() {
    document.getElementById('tools-btn').addEventListener('click', () => {
        const toolsPanel = document.getElementById('tools-panel');
        toolsPanel.classList.toggle('hidden');
    });

    document.getElementById('save-btn').addEventListener('click', saveCanvas);
    document.getElementById('load-btn').addEventListener('click', loadCanvas);

    document.getElementById('collapse-btn').addEventListener('click', collapseCanvas);

    document.getElementById('paint-water').addEventListener('click', () => setPaintColor('#0077be'));
    document.getElementById('paint-land').addEventListener('click', () => setPaintColor('#d2b55b'));
    document.getElementById('paint-forest').addEventListener('click', () => setPaintColor('#028A0F'));

    document.getElementById('brushSize').addEventListener('input', function (event) {
        brushSize = parseInt(event.target.value);
    });

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
            method: 'POST',
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

async function collapseCanvas() {
    const canvas = document.getElementById('paint-canvas');
    const width = canvas.width; // Set the desired width
    const height = canvas.height; // Set the desired height

    const rows = Math.ceil(height / TileSize);
    const cols = Math.ceil(width / TileSize);

    const paintedTiles = getPaintedTiles();

    console.log('There are ' + paintedTiles.length + ' provided painted tiles');

    const iterationSlider = document.getElementById('iterationSlider');
    let iterations = Number(iterationSlider.value);

    try {
        const response = await fetch('/generate', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({width: cols, height: rows, paintedTiles, iterations}),
        });


        if (response.ok) {
            const mapData = await response.json();


            // Render the generated map on the canvas
            const canvas = document.getElementById('paint-canvas');
            const ctx = canvas.getContext('2d');

            for (let y = 0; y < mapData.length; y++) {
                for (let x = 0; x < mapData[y].length; x++) {
                    const color = mapData[y][x].Color;
                    switch (color) {
                        case 0:
                            ctx.fillStyle = '#0077be'; // Water
                            break;
                        case 1:
                            ctx.fillStyle = '#7fc3e3'; // CoastalWater
                            break;
                        case 2:
                            ctx.fillStyle = '#d2b55b'; // Land
                            break;
                        case 3:
                            ctx.fillStyle = '#74B72E'; // Grass
                            break;
                        case 4:
                            ctx.fillStyle = '#028A0F'; // Forest
                            break;
                        default:
                            ctx.fillStyle = '#000000';
                            break;
                    }
                    ctx.fillRect(x * TileSize, y * TileSize, TileSize, TileSize); // 10x10 pixel size per tile
                }
            }
        } else {
            console.error('Failed to generate the map.');
        }
    } catch (error) {
        console.error('Error while generating the map:', error);
    }
}

function getColor(imageData) {
    for (let i = 0; i < imageData.data.length; i += 4) {
        const r = imageData.data[i];
        const g = imageData.data[i + 1];
        const b = imageData.data[i + 2];

        if (r === 0 && g === 119 && b === 190) {
            return 0; // Water
        } else if (r === 127 && g === 195 && b === 227) {
            return 1; // Coastal Water
        } else if (r === 210 && g === 181 && b === 91) {
            return 2; // Land
        } else if (r === 116 && g === 183 && b === 46) {
            return 3; // Grass
        } else if (r === 2 && g === 138 && b === 15) {
            return 4; // Forest
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
