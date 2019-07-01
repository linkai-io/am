package am

type Bag struct {
	Int32Vals   map[string][]int32
	Int64Vals   map[string][]int64
	Float32Vals map[string][]float32
	BoolVals    map[string][]bool
	StringVals  map[string][]string
}

func (b *Bag) AddBool(key string, value bool) {
	if b.BoolVals == nil {
		b.BoolVals = make(map[string][]bool)
	}

	if _, ok := b.BoolVals[key]; ok {
		b.BoolVals[key] = append(b.BoolVals[key], value)
	} else {
		b.BoolVals[key] = []bool{value}
	}
}

func (b *Bag) AddBools(key string, values []bool) {
	if b.BoolVals == nil {
		b.BoolVals = make(map[string][]bool)
	}

	if _, ok := b.BoolVals[key]; ok {
		b.BoolVals[key] = append(b.BoolVals[key], values...)
	} else {
		b.BoolVals[key] = values
	}
}

func (b *Bag) Bool(key string) (bool, bool) {
	if b.BoolVals != nil && len(b.BoolVals[key]) >= 1 {
		return b.BoolVals[key][0], true
	}
	return false, false
}

func (b *Bag) Bools(key string) ([]bool, bool) {
	if b.BoolVals != nil && len(b.BoolVals[key]) >= 1 {
		return b.BoolVals[key], true
	}
	return nil, false
}

func (b *Bag) AddInt32(key string, value int32) {
	if b.Int32Vals == nil {
		b.Int32Vals = make(map[string][]int32)
	}

	if _, ok := b.Int32Vals[key]; ok {
		b.Int32Vals[key] = append(b.Int32Vals[key], value)
	} else {
		b.Int32Vals[key] = []int32{value}
	}
}

func (b *Bag) AddInt32s(key string, values []int32) {
	if b.Int32Vals == nil {
		b.Int32Vals = make(map[string][]int32)
	}

	if _, ok := b.Int32Vals[key]; ok {
		b.Int32Vals[key] = append(b.Int32Vals[key], values...)
	} else {
		b.Int32Vals[key] = values
	}
}

func (b *Bag) Int32(key string) (int32, bool) {
	if b.Int32Vals != nil && len(b.Int32Vals[key]) >= 1 {
		return b.Int32Vals[key][0], true
	}
	return 0, false
}

func (b *Bag) Int32s(key string) ([]int32, bool) {
	if b.Int32Vals != nil && len(b.Int32Vals[key]) >= 1 {
		return b.Int32Vals[key], true
	}
	return nil, false
}

func (b *Bag) AddInt64(key string, value int64) {
	if b.Int64Vals == nil {
		b.Int64Vals = make(map[string][]int64)
	}

	if _, ok := b.Int64Vals[key]; ok {
		b.Int64Vals[key] = append(b.Int64Vals[key], value)
	} else {
		b.Int64Vals[key] = []int64{value}
	}
}

func (b *Bag) AddInt64s(key string, values []int64) {
	if b.Int64Vals == nil {
		b.Int64Vals = make(map[string][]int64)
	}

	if _, ok := b.Int64Vals[key]; ok {
		b.Int64Vals[key] = append(b.Int64Vals[key], values...)
	} else {
		b.Int64Vals[key] = values
	}
}

func (b *Bag) Int64(key string) (int64, bool) {
	if b.Int64Vals != nil && len(b.Int64Vals[key]) >= 1 {
		return b.Int64Vals[key][0], true
	}
	return 0, false
}

func (b *Bag) Int64s(key string) ([]int64, bool) {
	if b.Int64Vals != nil && len(b.Int64Vals[key]) >= 1 {
		return b.Int64Vals[key], true
	}
	return nil, false
}

func (b *Bag) AddFloat32(key string, value float32) {
	if b.Float32Vals == nil {
		b.Float32Vals = make(map[string][]float32)
	}

	if _, ok := b.Float32Vals[key]; ok {
		b.Float32Vals[key] = append(b.Float32Vals[key], value)
	} else {
		b.Float32Vals[key] = []float32{value}
	}
}

func (b *Bag) AddFloat32s(key string, values []float32) {
	if b.Float32Vals == nil {
		b.Float32Vals = make(map[string][]float32)
	}

	if _, ok := b.Float32Vals[key]; ok {
		b.Float32Vals[key] = append(b.Float32Vals[key], values...)
	} else {
		b.Float32Vals[key] = values
	}
}

func (b *Bag) Float32(key string) (float32, bool) {
	if b.Float32Vals != nil && len(b.Float32Vals[key]) >= 1 {
		return b.Float32Vals[key][0], true
	}
	return 0, false
}

func (b *Bag) Float32s(key string) ([]float32, bool) {
	if b.Float32Vals != nil && len(b.Float32Vals[key]) >= 1 {
		return b.Float32Vals[key], true
	}
	return nil, false
}

func (b *Bag) AddString(key, value string) {
	if b.StringVals == nil {
		b.StringVals = make(map[string][]string)
	}

	if _, ok := b.StringVals[key]; ok {
		b.StringVals[key] = append(b.StringVals[key], value)
	} else {
		b.StringVals[key] = []string{value}
	}
}

func (b *Bag) AddStrings(key string, values []string) {
	if b.StringVals == nil {
		b.StringVals = make(map[string][]string)
	}

	if _, ok := b.StringVals[key]; ok {
		b.StringVals[key] = append(b.StringVals[key], values...)
	} else {
		b.StringVals[key] = values
	}
}

func (b *Bag) String(key string) (string, bool) {
	if b.StringVals != nil && len(b.StringVals[key]) >= 1 {
		return b.StringVals[key][0], true
	}
	return "", false
}

func (b *Bag) Strings(key string) ([]string, bool) {
	if b.StringVals != nil && len(b.StringVals[key]) >= 1 {
		return b.StringVals[key], true
	}
	return nil, false
}
