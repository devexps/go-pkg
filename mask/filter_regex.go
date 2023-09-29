package mask

import (
	"regexp"

	"github.com/devexps/go-pkg/v2/mask/masker"
)

type regexFilter struct {
	regexList []regexp.Regexp
	maskType  masker.MType
}

// RegexFilter returns a Regex Filter.
func RegexFilter(regexPattern string, maskTypes ...masker.MType) Filter {
	f := &regexFilter{
		regexList: []regexp.Regexp{
			*regexp.MustCompile(regexPattern),
		},
	}
	if len(maskTypes) > 0 {
		f.maskType = maskTypes[0]
	}
	return f
}

// ReplaceString .
func (f *regexFilter) ReplaceString(s string) string {
	for _, p := range f.regexList {
		s = p.ReplaceAllString(s, maskerInstance.String(f.maskType, s))
	}
	return s
}

// MaskString .
func (f *regexFilter) MaskString(s string) string {
	return s
}

// ShouldMask .
func (f *regexFilter) ShouldMask(fieldName string, value interface{}, tag string) bool {
	return false
}
