package util

import (
	"fmt"
)

func BlockEmptyString(field string) func(val string) error {
	return func(val string) error {
		if len(val) == 0 {
			return fmt.Errorf("%s is required", field)
		}
		return nil
	}
}
