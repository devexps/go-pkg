package mask

import (
	"github.com/devexps/go-pkg/v2/mask/masker"
)

const tagKey = "mask"

type tagFilter struct {
	secureTags []string
	maskType   masker.MType
}

// TagsFilter returns a new filter instance with all available maskers
func TagsFilter() Filter {
	return TagFilter(maskerInstance.MarkTypes()...)
}

// TagFilter creates a new filter with multiple tags input
func TagFilter(tags ...masker.MType) Filter {
	if len(tags) == 0 {
		tags = []masker.MType{masker.MSecret}
	}
	var secureTags []string

	for _, tag := range tags {
		secureTags = append(secureTags, string(tag))
	}
	return &tagFilter{
		secureTags: secureTags,
	}
}

// ReplaceString .
func (f *tagFilter) ReplaceString(s string) string { return s }

// MaskString .
func (f *tagFilter) MaskString(s string) string {
	return maskerInstance.String(f.maskType, s)
}

// ShouldMask .
func (f *tagFilter) ShouldMask(fieldName string, value interface{}, tag string) bool {
	for _, stag := range f.secureTags {
		if stag == tag {
			f.maskType = masker.MType(tag)
			return true
		}
	}
	return false
}
