/**
 * Finds all contiguous clusters of same-colored tiles in a grid.
 * @param {number[][]} grid - The 2D array of tile indices.
 * @returns {Array<{tileType: number, size: number}>} An array of cluster objects.
 */
export function getClusters(grid) {
    const rows = grid.length;
    const cols = grid[0].length;
    const seen = Array.from({length: rows}, () => Array(cols).fill(false));
    const clusters = [];

    for (let y = 0; y < rows; y++) {
        for (let x = 0; x < cols; x++) {
            if (seen[y][x]) continue;
            const type = grid[y][x];
            let size = 0;
            const stack = [[x, y]];
            seen[y][x] = true;

            while (stack.length) {
                const [cx, cy] = stack.pop();
                size++;
                [[1, 0], [-1, 0], [0, 1], [0, -1]].forEach(([dx, dy]) => {
                    const nx = cx + dx;
                    const ny = cy + dy;
                    if (nx >= 0 && nx < cols && ny >= 0 && ny < rows && !seen[ny][nx] && grid[ny][nx] === type) {
                        seen[ny][nx] = true;
                        stack.push([nx, ny]);
                    }
                });
            }
            clusters.push({tileType: type, size});
        }
    }
    return clusters;
}