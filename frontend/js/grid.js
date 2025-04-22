const GridSize = 1000;
const TileSize = 10;


function initGrid() {

    const gridCanvas = document.getElementById('grid-canvas');
    gridCanvas.width = GridSize;
    gridCanvas.height = GridSize;

    const paintCanvas = document.getElementById('paint-canvas');
    paintCanvas.width = GridSize;
    paintCanvas.height = GridSize;

    drawGrid();
}


function drawGrid() {
    const canvas = document.getElementById('grid-canvas');
    const ctx = canvas.getContext('2d');

    ctx.clearRect(0, 0, canvas.width, canvas.height);

    drawSquareGrid(ctx, canvas.width, canvas.height);
}

function drawSquareGrid(ctx, width, height) {
    const gridSize = GridSize;

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
