let paintColor = null;

window.setPaintColor = function (color) {
    paintColor = color;
}

function initPainting() {
    const canvas = document.getElementById('paint-canvas');
    const ctx = canvas.getContext('2d');

    canvas.addEventListener('mousedown', startPainting);
    canvas.addEventListener('mouseup', stopPainting);
    canvas.addEventListener('mousemove', paint);

    let painting = false;

    function startPainting(event) {
        if (paintColor) {
            painting = true;
            paint(event);
        }
    }

    function stopPainting() {
        painting = false;
    }

    function paint(event) {
        if (!painting) return;

        const canvasRect = canvas.getBoundingClientRect();
        const adjustedX = Math.floor((event.clientX - canvasRect.left) / TileSize) * TileSize;
        const adjustedY = Math.floor((event.clientY - canvasRect.top) / TileSize) * TileSize;

        ctx.fillStyle = paintColor;
        for (let y = 0; y < brushSize; y++) {
            for (let x = 0; x < brushSize; x++) {
                ctx.fillRect(adjustedX + x * TileSize, adjustedY + y * TileSize, TileSize, TileSize);
            }
        }
    }
}

document.addEventListener('DOMContentLoaded', initPainting);
