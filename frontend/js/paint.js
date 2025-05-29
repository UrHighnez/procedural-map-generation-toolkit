import {TileSize} from "./grid.js";

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
        const mouseX = event.clientX - canvasRect.left;
        const mouseY = event.clientY - canvasRect.top;

        const centerCellX = Math.floor(mouseX / TileSize);
        const centerCellY = Math.floor(mouseY / TileSize);

        // Calculate offset
        const halfBrush = Math.floor(brushSize / 2);

        ctx.fillStyle = paintColor;
        for (let y = 0; y < brushSize; y++) {
            for (let x = 0; x < brushSize; x++) {
                ctx.fillRect((centerCellX + x - halfBrush) * TileSize, (centerCellY + y - halfBrush) * TileSize, TileSize, TileSize);
            }
        }
    }
}

document.addEventListener('DOMContentLoaded', initPainting);
