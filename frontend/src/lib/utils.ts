export function getPointToRectProximity(x: number, y: number, rect: DOMRect) {
	const dx = Math.max(rect.left - x, 0, x - rect.right + 1);
	const dy = Math.max(rect.top - y, 0, y - rect.bottom + 1);
	return Math.sqrt(dx * dx + dy * dy);
}
