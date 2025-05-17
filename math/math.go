package beam_math

func ManhattanDistance(x1, y1, x2, y2 int) int {
	return Abs(x1-x2) + Abs(y1-y2)
}

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func Sign(x int) int {
	if x < 0 {
		return -1
	}
	if x > 0 {
		return 1
	}
	return 0
}

func GetTriangleWeight(floor, start, end int) float32 {
	if floor < start || floor > end {
		return 0
	}

	mid := (start + end) / 2
	width := float32(end - start)
	if floor <= mid {
		return 2 * float32(floor-start) / width
	}
	return 2 * float32(end-floor) / width
}
