export async function getColors() {
    const res = await fetch('/colors');
    if (!res.ok) throw new Error('Failed to load colors from server');
    return res.json();
}

export async function saveCanvas(canvas) {
    const imageData = canvas.toDataURL('image/png');
    const response = await fetch('/save', {
        method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({imageData}),
    });
    if (!response.ok) {
        throw new Error('Failed to save the canvas image.');
    }
    console.log('Canvas image saved successfully.');
}

export async function loadCanvasTo(canvas) {
    const response = await fetch('/load');
    if (!response.ok) {
        throw new Error('Failed to load map images.');
    }
    const imageFiles = await response.json();
    const selectedFile = prompt('Choose a map file to load:', imageFiles.join(', '));
    if (selectedFile) {
        const img = new Image();
        img.onload = () => {
            const ctx = canvas.getContext('2d');
            ctx.clearRect(0, 0, canvas.width, canvas.height);
            ctx.drawImage(img, 0, 0);
        };
        img.src = `/maps/${selectedFile}`;
    }
}

export async function generate(params) {
    const response = await fetch('/generate', {
        method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify(params),
    });

    if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`Failed to generate the map: ${response.status} ${errorText}`);
    }
    return response.json();
}