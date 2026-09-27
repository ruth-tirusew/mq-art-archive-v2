export type DiffLine = { type: 'equal' | 'insert' | 'delete'; value: string };

// Line-level LCS diff. Fine for article-length text; not meant for huge documents (O(n*m)).
export function diffLines(oldText: string, newText: string): DiffLine[] {
	const oldLines = oldText.split('\n');
	const newLines = newText.split('\n');
	const n = oldLines.length;
	const m = newLines.length;

	const dp: number[][] = Array.from({ length: n + 1 }, () => new Array<number>(m + 1).fill(0));
	for (let i = n - 1; i >= 0; i--) {
		for (let j = m - 1; j >= 0; j--) {
			dp[i][j] =
				oldLines[i] === newLines[j] ? dp[i + 1][j + 1] + 1 : Math.max(dp[i + 1][j], dp[i][j + 1]);
		}
	}

	const result: DiffLine[] = [];
	let i = 0;
	let j = 0;
	while (i < n && j < m) {
		if (oldLines[i] === newLines[j]) {
			result.push({ type: 'equal', value: oldLines[i] });
			i++;
			j++;
		} else if (dp[i + 1][j] >= dp[i][j + 1]) {
			result.push({ type: 'delete', value: oldLines[i] });
			i++;
		} else {
			result.push({ type: 'insert', value: newLines[j] });
			j++;
		}
	}
	while (i < n) {
		result.push({ type: 'delete', value: oldLines[i] });
		i++;
	}
	while (j < m) {
		result.push({ type: 'insert', value: newLines[j] });
		j++;
	}
	return result;
}
