package mask

import (
	"reflect"

	"github.com/devexps/go-pkg/v2/mask/masker"
)

type typeFilter struct {
	target   reflect.Type
	maskType masker.MType
}

// TypeFilter returns Type Filter with custom masking type.
func TypeFilter(t interface{}, maskTypes ...masker.MType) *typeFilter {
	f := &typeFilter{
		target: reflect.TypeOf(t),
	}
	if len(maskTypes) > 0 {
		f.maskType = maskTypes[0]
	}
	return f
}

// ReplaceString .
func (f *typeFilter) ReplaceString(s string) string {
	return s
}

// MaskString .
func (f *typeFilter) MaskString(s string) string {
	return maskerInstance.String(f.maskType, s)
}

// ShouldMask .
func (f *typeFilter) ShouldMask(fieldName string, value interface{}, tag string) bool {
	return f.target == reflect.TypeOf(value)
}
