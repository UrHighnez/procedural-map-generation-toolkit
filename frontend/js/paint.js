let paintColor = null;
let isGreyscale = false;
let originalImageData = null;


window.setPaintColor = function (color) {
    paintColor = color;
}

window.toggleGreyscale = function (greyscaleOn) {
    const canvas = document.getElementById('paint-canvas');
    const ctx = canvas.getContext('2d');

    if (greyscaleOn) {
        // Save image data (Only first time)
        if (!originalImageData) {
            originalImageData = ctx.getImageData(0, 0, canvas.width, canvas.height);
        }
        const imgData = ctx.getImageData(0, 0, canvas.width, canvas.height);
        for (let i = 0; i < imgData.data.length; i += 4) {
            const avg = 0.299 * imgData.data[i] + 0.587 * imgData.data[i + 1] + 0.114 * imgData.data[i + 2];
            imgData.data[i] = imgData.data[i + 1] = imgData.data[i + 2] = avg;
        }
        ctx.putImageData(imgData, 0, 0);
        isGreyscale = true;
    } else if (originalImageData) {
        // Recover original image data
        ctx.putImageData(originalImageData, 0, 0);
        isGreyscale = false;
    }
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
                ctx.fillRect(
                    (centerCellX + x - halfBrush) * TileSize,
                    (centerCellY + y - halfBrush) * TileSize,
                    TileSize,
                    TileSize
                );
            }
        }
    }

}

document.addEventListener('DOMContentLoaded', initPainting);
