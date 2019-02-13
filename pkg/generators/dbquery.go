package generators

import "fmt"

func AppendConditionalQuery(query, prefix *string, filter string, value interface{}, args *[]interface{}, i *int) {
	filter = fmt.Sprintf(filter, *i)
	*query += fmt.Sprintf("%s%s", *prefix, filter)
	*i++
	*args = append(*args, value)
	if *prefix == "" {
		*prefix = " and "
	}
}
