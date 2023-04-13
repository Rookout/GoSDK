package types

import (
	"regexp"
)

type FieldFilterType int

type FieldFilter struct {
	FilterType FieldFilterType
	Pattern    *regexp.Regexp
	Whitelist  bool
}
