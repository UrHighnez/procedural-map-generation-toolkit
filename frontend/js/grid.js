let gridType = 'square'; // Default grid type

function initGrid() {
    // Call resizeGridCanvas when the page is loaded
    resizeGridCanvas();

    // Call resizeGridCanvas when the window is resized
    window.addEventListener('resize', () => {
        resizeGridCanvas();
    });

    resizeBackgroundCanvas();
}

function resizeGridCanvas() {
    const gridCanvas = document.getElementById('grid-canvas');
    gridCanvas.width = window.innerWidth;
    gridCanvas.height = window.innerHeight;

    resizePaintCanvas();
    resizeBackgroundCanvas();
    drawGrid();
}

function resizePaintCanvas() {
    const paintCanvas = document.getElementById('paint-canvas');
    paintCanvas.width = window.innerWidth;
    paintCanvas.height = window.innerHeight;
}

function resizeBackgroundCanvas() {
    const backgroundCanvas = document.getElementById('background-canvas');
    backgroundCanvas.width = window.innerWidth;
    backgroundCanvas.height = window.innerHeight;
}

function toggleGrid() {
    gridType = gridType === 'square' ? 'hexagonal' : 'square';
    drawGrid();
}

function drawGrid() {
    const canvas = document.getElementById('grid-canvas');
    const ctx = canvas.getContext('2d');

    ctx.clearRect(0, 0, canvas.width, canvas.height);

    if (gridType === 'square') {
        drawSquareGrid(ctx, canvas.width, canvas.height);
    } else {
        drawHexagonalGrid(ctx, canvas.width, canvas.height);
    }
}

function drawSquareGrid(ctx, width, height) {
    const gridSize = 20;

    ctx.strokeStyle = 'black';
    ctx.lineWidth = 0.5;

    for (let x = 0; x <= width; x += gridSize) {
        ctx.beginPath();
        ctx.moveTo(x, 0);
        ctx.lineTo(x, height);
        ctx.stroke();
    }

    for (let y = 0; y <= height; y += gridSize) {
        ctx.beginPath();
        ctx.moveTo(0, y);
        ctx.lineTo(width, y);
        ctx.stroke();
    }
}

function drawHexagonalGrid(ctx, width, height) {
    const hexSize = 13.33;
    const hexHeight = hexSize * 2;
    const hexWidth = Math.sqrt(3) * hexSize;
    const hexHalfHeight = hexSize / 2;
    const hexHalfWidth = hexWidth / 2;
    const verticalSpacing = hexSize * 1.5;
    const horizontalSpacing = hexWidth;

    ctx.strokeStyle = 'black';
    ctx.lineWidth = 0.5;

    const drawHexagon = (x, y) => {
        ctx.beginPath();
        ctx.moveTo(x + hexHalfWidth, y);
        ctx.lineTo(x + hexWidth, y + hexHalfHeight);
        ctx.lineTo(x + hexWidth, y + hexHeight - hexHalfHeight);
        ctx.lineTo(x + hexHalfWidth, y + hexHeight);
        ctx.lineTo(x, y + hexHeight - hexHalfHeight);
        ctx.lineTo(x, y + hexHalfHeight);
        ctx.closePath();
        ctx.stroke();
    };

    for (let y = 0, rowIndex = 0; y < height + hexHeight; y += verticalSpacing, rowIndex++) {
        for (let x = 0; x < width + hexWidth; x += horizontalSpacing) {
            let offsetX = (rowIndex % 2 === 0) ? 0 : hexHalfWidth;
            drawHexagon(x + offsetX, y);
        }
    }
}
