async function exportMetricAsImage(elementId, filenamePrefix) {
    const containerElement = document.getElementById(elementId);
    if (!containerElement) {
        console.error('Container element for export not found:', elementId);
        alert('Could not find the container to export.');
        return;
    }

    let elementToCapture = containerElement;

    if (containerElement.children.length === 1 && containerElement.children[0].tagName === 'TABLE') {
        elementToCapture = containerElement.children[0];
        console.log(`Capturing specific table child for ${elementId}`);
    } else {
        console.log(`Capturing container element for ${elementId}`);
    }

    // Visible if in a scrollable container
    const parentBox = elementToCapture.closest('.metric-display-box') || containerElement.closest('.metric-item');
    let originalScrollTop, originalScrollLeft;
    if (parentBox && parentBox.scrollHeight > parentBox.clientHeight || parentBox.scrollWidth > parentBox.clientWidth) {
        originalScrollTop = parentBox.scrollTop;
        originalScrollLeft = parentBox.scrollLeft;
        if (elementToCapture !== parentBox) {
            elementToCapture.scrollIntoView({behavior: 'instant', block: 'nearest', inline: 'nearest'});
        } else {
            parentBox.scrollTop = 0;
            parentBox.scrollLeft = 0;
        }
    }


    try {
        const canvas = await html2canvas(elementToCapture, {
            logging: false,
            useCORS: true,
            scale: 2,
            backgroundColor: window.getComputedStyle(elementToCapture.closest('.metric-item') || elementToCapture).backgroundColor || '#ffffff',
        });

        const imageDataUrl = canvas.toDataURL('image/png');
        const link = document.createElement('a');
        link.href = imageDataUrl;
        link.download = `${filenamePrefix}_${elementId}.png`;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);

    } catch (error) {
        console.error(`Error exporting ${elementId}:`, error);
        alert(`Could not export ${elementId} visualization.`);
    } finally {
        if (parentBox && (originalScrollTop !== undefined || originalScrollLeft !== undefined)) {
            if (originalScrollTop !== undefined) parentBox.scrollTop = originalScrollTop;
            if (originalScrollLeft !== undefined) parentBox.scrollLeft = originalScrollLeft;
        }
    }
}

export function initExportButtons() {

    document.querySelectorAll('.export-metric-btn').forEach(button => {
        const targetId = button.dataset.exportId;
        button.addEventListener('click', () => {
            let filenamePrefix = targetId.charAt(0).toUpperCase() + targetId.slice(1);
            if (targetId === "exportId") filenamePrefix = "Metric";

            exportMetricAsImage(targetId, filenamePrefix);
        });
    });
}