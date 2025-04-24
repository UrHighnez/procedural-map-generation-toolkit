const GridSize = 500;
const TileSize = 20;


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

    ctx.clearRect(0, 0, GridSize, GridSize);

    drawSquareGrid(ctx, GridSize, GridSize);
}

function drawSquareGrid(ctx, width, height) {

    ctx.strokeStyle = 'black';
    ctx.lineWidth = 0.5;

    for (let x = 0; x <= width; x += TileSize) {
        ctx.beginPath();
        ctx.moveTo(x, 0);
        ctx.lineTo(x, height);
        ctx.stroke();
    }

    for (let y = 0; y <= height; y += TileSize) {
        ctx.beginPath();
        ctx.moveTo(0, y);
        ctx.lineTo(width, y);
        ctx.stroke();
    }
}
