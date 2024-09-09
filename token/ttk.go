package token

import (
	"regexp"
)

var (
	reg = regexp.MustCompile(`tkk:'(\d+\.\d+)'`)
)

func MustGetTKK() string {
	return "444000.1270171236"
	// return "445767.3058494238"
	// return "445678.1618007056"
	// return "444444.1050258596"
	// return "445111.1710346305"
}
