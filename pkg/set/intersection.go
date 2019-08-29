package set

func StrIntersection(a, b []string) []string {
	if a == nil || b == nil || len(a) == 0 || len(b) == 0 {
		return []string{}
	}

	m := make(map[string]bool)
	c := make([]string, 0)

	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		if _, ok := m[item]; ok {
			c = append(c, item)
		}
	}

	return c
}

func Int32Intersection(a, b []int32) []int32 {
	if a == nil || b == nil || len(a) == 0 || len(b) == 0 {
		return []int32{}
	}
	m := make(map[int32]bool)
	c := make([]int32, 0)

	for _, item := range a {
		m[item] = true
	}

	for _, item := range b {
		if _, ok := m[item]; ok {
			c = append(c, item)
		}
	}

	return c
}
