package mask

import (
	"strings"

	"github.com/devexps/go-pkg/v2/mask/masker"
)

type fieldPrefixFilter struct {
	prefix   string
	maskType masker.MType
}

// FieldPrefixFilter returns a Field Prefix Filter.
func FieldPrefixFilter(prefix string, maskTypes ...masker.MType) Filter {
	f := &fieldPrefixFilter{
		prefix: prefix,
	}
	if len(maskTypes) > 0 {
		f.maskType = maskTypes[0]
	}
	return f
}

// ReplaceString .
func (f *fieldPrefixFilter) ReplaceString(s string) string {
	return s
}

// MaskString .
func (f *fieldPrefixFilter) MaskString(s string) string {
	return maskerInstance.String(f.maskType, s)
}

// ShouldMask .
func (f *fieldPrefixFilter) ShouldMask(fieldName string, value interface{}, tag string) bool {
	return strings.HasPrefix(fieldName, f.prefix)
}
