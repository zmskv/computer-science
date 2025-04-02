package main

func isSymmetric(points [][2]int) bool {
	maxVal := -10000
	minVal := 10000
	for i := 0; i < len(points); i++ {
		if points[i][0] > maxVal {
			maxVal = points[i][0]
		}
	}

	for j := 0; j < len(points); j++ {
		if points[j][0] < minVal {
			minVal = points[j][0]
		}
	}

	var midPoint [2]int

	for k := 0; k < len(points); k++ {
		if maxVal+minVal == 0 && points[k][0] == 0 {
			midPoint = points[k]
			break
		} else if (maxVal+minVal)/2 == points[k][0] {
			midPoint = points[k]
			break
		} else {
			midPoint = [2]int{}
		}
	}

	if len(midPoint) == 0 {
		return false
	}

	findX := midPoint[0]
	m := make(map[[2]int]bool, 0)
	for _, p := range points {
		m[p] = false
	}

	for _, p := range points {
		if _, ok := m[[2]int{2*findX - p[0], p[1]}]; ok {
			m[[2]int{p[0], p[1]}] = true
		}
	}

	for _, p := range points {
		if !m[p] {
			return false
		}
	}

	return true
}

// асимптотика O(n), где n - количество точек
// сложность по памяти O(n), где n - количество точек в мапе
