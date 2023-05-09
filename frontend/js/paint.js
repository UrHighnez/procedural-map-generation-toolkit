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
        const adjustedX = event.clientX - canvasRect.left;
        const adjustedY = event.clientY - canvasRect.top;

        ctx.fillStyle = paintColor;
        ctx.fillRect(Math.floor(adjustedX / 30) * 30, Math.floor(adjustedY / 30) * 30, 30, 30);
    }
}

document.addEventListener('DOMContentLoaded', initPainting);
