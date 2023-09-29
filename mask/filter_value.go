package mask

import (
	"strings"

	"github.com/devexps/go-pkg/v2/mask/masker"
)

type valueFilter struct {
	value    string
	maskType masker.MType
}

// ValueFilter returns a Value Filter with custom masking type.
func ValueFilter(target string, maskTypes ...masker.MType) Filter {
	f := &valueFilter{
		value: target,
	}
	if len(maskTypes) > 0 {
		f.maskType = maskTypes[0]
	}
	return f
}

// ReplaceString .
func (f *valueFilter) ReplaceString(s string) string {
	return strings.ReplaceAll(s, f.value, maskerInstance.String(f.maskType, s))
}

// MaskString .
func (f *valueFilter) MaskString(s string) string {
	return s
}

// ShouldMask .
func (f *valueFilter) ShouldMask(fieldName string, value interface{}, tag string) bool {
	return false
}
