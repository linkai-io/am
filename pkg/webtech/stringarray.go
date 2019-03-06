package webtech

import (
	"encoding/json"
	"fmt"
)

// StringArray type is a wrapper for []string for use in unmarshalling the apps.json
type StringArray []string

// UnmarshalJSON is a custom unmarshaler for handling bogus apps.json types from wappalyzer
func (t *StringArray) UnmarshalJSON(data []byte) error {
	var s string
	var sa []string
	var na []int

	if err := json.Unmarshal(data, &s); err != nil {
		if err := json.Unmarshal(data, &na); err == nil {
			// not a string, so maybe []int?
			*t = make(StringArray, len(na))

			for i, number := range na {
				(*t)[i] = fmt.Sprintf("%d", number)
			}

			return nil
		} else if err := json.Unmarshal(data, &sa); err == nil {
			// not a string, so maybe []string?
			*t = sa
			return nil
		}
		fmt.Println(string(data))
		return err
	}
	*t = StringArray{s}
	return nil
}
