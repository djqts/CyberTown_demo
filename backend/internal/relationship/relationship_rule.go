package relationship

func ClampAffinity(val int) int {
	if val < 0 {
		return 0
	}
	if val > 100 {
		return 100
	}
	return val
}
