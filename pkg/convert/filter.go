package convert

import (
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/protocservices/prototypes"
)

func DomainToFilterTypes(in *am.FilterType) *prototypes.FilterType {
	filter := &prototypes.FilterType{}
	if in.BoolFilters != nil {
		filter.BoolFilters = make(map[string]*prototypes.BoolFilter)
		for k, v := range in.BoolFilters {
			if v == nil {
				continue
			}
			filter.BoolFilters[k] = &prototypes.BoolFilter{Value: v}
		}
	}
	if in.Int32Filters != nil {
		filter.Int32Filters = make(map[string]*prototypes.Int32Filter)
		for k, v := range in.Int32Filters {
			if v == nil {
				continue
			}
			filter.Int32Filters[k] = &prototypes.Int32Filter{Value: v}
		}
	}
	if in.Int64Filters != nil {
		filter.Int64Filters = make(map[string]*prototypes.Int64Filter)
		for k, v := range in.Int64Filters {
			if v == nil {
				continue
			}
			filter.Int64Filters[k] = &prototypes.Int64Filter{Value: v}
		}
	}
	if in.Float32Filters != nil {
		filter.FloatFilters = make(map[string]*prototypes.FloatFilter)
		for k, v := range in.Float32Filters {
			if v == nil {
				continue
			}
			filter.FloatFilters[k] = &prototypes.FloatFilter{Value: v}
		}
	}

	if in.StringFilters != nil {
		filter.StringFilters = make(map[string]*prototypes.StringFilter)
		for k, v := range in.StringFilters {
			if v == nil {
				continue
			}
			filter.StringFilters[k] = &prototypes.StringFilter{Value: v}
		}
	}
	return filter
}

func FilterTypesToDomain(in *prototypes.FilterType) *am.FilterType {
	filter := &am.FilterType{}
	if in.BoolFilters != nil {
		filter.BoolFilters = make(map[string][]bool)
		for k, v := range in.BoolFilters {
			if v == nil {
				continue
			}
			filter.BoolFilters[k] = v.Value
		}
	}
	if in.Int32Filters != nil {
		filter.Int32Filters = make(map[string][]int32)
		for k, v := range in.Int32Filters {
			if v == nil {
				continue
			}
			filter.Int32Filters[k] = v.Value
		}
	}
	if in.Int64Filters != nil {
		filter.Int64Filters = make(map[string][]int64)
		for k, v := range in.Int64Filters {
			if v == nil {
				continue
			}
			filter.Int64Filters[k] = v.Value
		}
	}
	if in.FloatFilters != nil {
		filter.Float32Filters = make(map[string][]float32)
		for k, v := range in.FloatFilters {
			if v == nil {
				continue
			}
			filter.Float32Filters[k] = v.Value
		}
	}
	if in.StringFilters != nil {
		filter.StringFilters = make(map[string][]string)
		for k, v := range in.StringFilters {
			if v == nil {
				continue
			}
			filter.StringFilters[k] = v.Value
		}
	}
	return filter
}
