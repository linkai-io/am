package am

type FilterType struct {
	Int32Filters   map[string][]int32
	Int64Filters   map[string][]int64
	Float32Filters map[string][]float32
	BoolFilters    map[string][]bool
	StringFilters  map[string][]string
}

func (f *FilterType) AddBool(key string, value bool) {
	if f.BoolFilters == nil {
		f.BoolFilters = make(map[string][]bool)
	}

	if _, ok := f.BoolFilters[key]; ok {
		f.BoolFilters[key] = append(f.BoolFilters[key], value)
	} else {
		f.BoolFilters[key] = []bool{value}
	}
}

func (f *FilterType) AddBools(key string, values []bool) {
	if f.BoolFilters == nil {
		f.BoolFilters = make(map[string][]bool)
	}

	if _, ok := f.BoolFilters[key]; ok {
		f.BoolFilters[key] = append(f.BoolFilters[key], values...)
	} else {
		f.BoolFilters[key] = values
	}
}

func (f *FilterType) Bool(key string) (bool, bool) {
	if f.BoolFilters != nil && len(f.BoolFilters[key]) >= 1 {
		return f.BoolFilters[key][0], true
	}
	return false, false
}

func (f *FilterType) Bools(key string) ([]bool, bool) {
	if f.BoolFilters != nil && len(f.BoolFilters[key]) >= 1 {
		return f.BoolFilters[key], true
	}
	return nil, false
}

func (f *FilterType) AddInt32(key string, value int32) {
	if f.Int32Filters == nil {
		f.Int32Filters = make(map[string][]int32)
	}

	if _, ok := f.Int32Filters[key]; ok {
		f.Int32Filters[key] = append(f.Int32Filters[key], value)
	} else {
		f.Int32Filters[key] = []int32{value}
	}
}

func (f *FilterType) AddInt32s(key string, values []int32) {
	if f.Int32Filters == nil {
		f.Int32Filters = make(map[string][]int32)
	}

	if _, ok := f.Int32Filters[key]; ok {
		f.Int32Filters[key] = append(f.Int32Filters[key], values...)
	} else {
		f.Int32Filters[key] = values
	}
}

func (f *FilterType) Int32(key string) (int32, bool) {
	if f.Int32Filters != nil && len(f.Int32Filters[key]) >= 1 {
		return f.Int32Filters[key][0], true
	}
	return 0, false
}

func (f *FilterType) Int32s(key string) ([]int32, bool) {
	if f.Int32Filters != nil && len(f.Int32Filters[key]) >= 1 {
		return f.Int32Filters[key], true
	}
	return nil, false
}

func (f *FilterType) AddInt64(key string, value int64) {
	if f.Int64Filters == nil {
		f.Int64Filters = make(map[string][]int64)
	}

	if _, ok := f.Int64Filters[key]; ok {
		f.Int64Filters[key] = append(f.Int64Filters[key], value)
	} else {
		f.Int64Filters[key] = []int64{value}
	}
}

func (f *FilterType) AddInt64s(key string, values []int64) {
	if f.Int64Filters == nil {
		f.Int64Filters = make(map[string][]int64)
	}

	if _, ok := f.Int64Filters[key]; ok {
		f.Int64Filters[key] = append(f.Int64Filters[key], values...)
	} else {
		f.Int64Filters[key] = values
	}
}

func (f *FilterType) Int64(key string) (int64, bool) {
	if f.Int64Filters != nil && len(f.Int64Filters[key]) >= 1 {
		return f.Int64Filters[key][0], true
	}
	return 0, false
}

func (f *FilterType) Int64s(key string) ([]int64, bool) {
	if f.Int64Filters != nil && len(f.Int64Filters[key]) >= 1 {
		return f.Int64Filters[key], true
	}
	return nil, false
}

func (f *FilterType) AddFloat32(key string, value float32) {
	if f.Float32Filters == nil {
		f.Float32Filters = make(map[string][]float32)
	}

	if _, ok := f.Float32Filters[key]; ok {
		f.Float32Filters[key] = append(f.Float32Filters[key], value)
	} else {
		f.Float32Filters[key] = []float32{value}
	}
}

func (f *FilterType) AddFloat32s(key string, values []float32) {
	if f.Float32Filters == nil {
		f.Float32Filters = make(map[string][]float32)
	}

	if _, ok := f.Float32Filters[key]; ok {
		f.Float32Filters[key] = append(f.Float32Filters[key], values...)
	} else {
		f.Float32Filters[key] = values
	}
}

func (f *FilterType) Float32(key string) (float32, bool) {
	if f.Float32Filters != nil && len(f.Float32Filters[key]) >= 1 {
		return f.Float32Filters[key][0], true
	}
	return 0, false
}

func (f *FilterType) Float32s(key string) ([]float32, bool) {
	if f.Float32Filters != nil && len(f.Float32Filters[key]) >= 1 {
		return f.Float32Filters[key], true
	}
	return nil, false
}

func (f *FilterType) AddString(key, value string) {
	if f.StringFilters == nil {
		f.StringFilters = make(map[string][]string)
	}

	if _, ok := f.StringFilters[key]; ok {
		f.StringFilters[key] = append(f.StringFilters[key], value)
	} else {
		f.StringFilters[key] = []string{value}
	}
}

func (f *FilterType) AddStrings(key string, values []string) {
	if f.StringFilters == nil {
		f.StringFilters = make(map[string][]string)
	}

	if _, ok := f.StringFilters[key]; ok {
		f.StringFilters[key] = append(f.StringFilters[key], values...)
	} else {
		f.StringFilters[key] = values
	}
}

func (f *FilterType) String(key string) (string, bool) {
	if f.StringFilters != nil && len(f.StringFilters[key]) >= 1 {
		return f.StringFilters[key][0], true
	}
	return "", false
}

func (f *FilterType) Strings(key string) ([]string, bool) {
	if f.StringFilters != nil && len(f.StringFilters[key]) >= 1 {
		return f.StringFilters[key], true
	}
	return nil, false
}
