package convert

import (
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/protocservices/prototypes"
)

func DomainToBags(in *am.Bag) *prototypes.Bag {
	Val := &prototypes.Bag{}
	if in.BoolVals != nil {
		Val.BoolVals = make(map[string]*prototypes.BoolVal)
		for k, v := range in.BoolVals {
			if v == nil {
				continue
			}
			Val.BoolVals[k] = &prototypes.BoolVal{Value: v}
		}
	}
	if in.Int32Vals != nil {
		Val.Int32Vals = make(map[string]*prototypes.Int32Val)
		for k, v := range in.Int32Vals {
			if v == nil {
				continue
			}
			Val.Int32Vals[k] = &prototypes.Int32Val{Value: v}
		}
	}
	if in.Int64Vals != nil {
		Val.Int64Vals = make(map[string]*prototypes.Int64Val)
		for k, v := range in.Int64Vals {
			if v == nil {
				continue
			}
			Val.Int64Vals[k] = &prototypes.Int64Val{Value: v}
		}
	}
	if in.Float32Vals != nil {
		Val.FloatVals = make(map[string]*prototypes.FloatVal)
		for k, v := range in.Float32Vals {
			if v == nil {
				continue
			}
			Val.FloatVals[k] = &prototypes.FloatVal{Value: v}
		}
	}

	if in.StringVals != nil {
		Val.StringVals = make(map[string]*prototypes.StringVal)
		for k, v := range in.StringVals {
			if v == nil {
				continue
			}
			Val.StringVals[k] = &prototypes.StringVal{Value: v}
		}
	}
	return Val
}

func BagsToDomain(in *prototypes.Bag) *am.Bag {
	Val := &am.Bag{}
	if in.BoolVals != nil {
		Val.BoolVals = make(map[string][]bool)
		for k, v := range in.BoolVals {
			if v == nil {
				continue
			}
			Val.BoolVals[k] = v.Value
		}
	}
	if in.Int32Vals != nil {
		Val.Int32Vals = make(map[string][]int32)
		for k, v := range in.Int32Vals {
			if v == nil {
				continue
			}
			Val.Int32Vals[k] = v.Value
		}
	}
	if in.Int64Vals != nil {
		Val.Int64Vals = make(map[string][]int64)
		for k, v := range in.Int64Vals {
			if v == nil {
				continue
			}
			Val.Int64Vals[k] = v.Value
		}
	}
	if in.FloatVals != nil {
		Val.Float32Vals = make(map[string][]float32)
		for k, v := range in.FloatVals {
			if v == nil {
				continue
			}
			Val.Float32Vals[k] = v.Value
		}
	}
	if in.StringVals != nil {
		Val.StringVals = make(map[string][]string)
		for k, v := range in.StringVals {
			if v == nil {
				continue
			}
			Val.StringVals[k] = v.Value
		}
	}
	return Val
}
