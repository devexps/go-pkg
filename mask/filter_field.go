package mask

import (
	"github.com/devexps/go-pkg/v2/mask/masker"
)

type fieldFilter struct {
	fieldName string
	maskType  masker.MType
}

// FieldFilter returns a Field Filter.
func FieldFilter(fieldName string, maskTypes ...masker.MType) Filter {
	f := &fieldFilter{
		fieldName: fieldName,
	}
	if len(maskTypes) > 0 {
		f.maskType = maskTypes[0]
	}
	return f
}

// ReplaceString .
func (f *fieldFilter) ReplaceString(s string) string {
	return s
}

// MaskString .
func (f *fieldFilter) MaskString(s string) string {
	return maskerInstance.String(f.maskType, s)
}

// ShouldMask .
func (f *fieldFilter) ShouldMask(fieldName string, value interface{}, tag string) bool {
	return f.fieldName == fieldName
}
