package mask

type Filter interface {
	// ReplaceString is called when checking string type.
	// The argument is the value to be checked, and the return value should be the value to be replaced.
	// If nothing needs to be done, the method should return the argument as is.
	// This method is intended for the case where you want to hide a part of a string.
	ReplaceString(s string) string

	// MaskString is called when checking field, prefix type and tag type.
	// The return value is to be replaced.
	// This method is intended for the case where you want to hide a part of a string.
	MaskString(s string) string

	// ShouldMask is called for all values to be checked.
	// The field name of the value to be checked, and tag value if the structure has `zlog` tag will be passed as arguments.
	// If the return value is false, nothing is done; if it is true, the entire field is hidden.
	// Hidden values will be replaced with the value "[filtered]" if string type.
	ShouldMask(fieldName string, value interface{}, tag string) bool
}

type Filters []Filter

// ReplaceString .
func (fs Filters) ReplaceString(s string) string {
	for _, f := range fs {
		s = f.ReplaceString(s)
	}
	return s
}

// MaskString .
func (fs Filters) MaskString(s string) string {
	for _, f := range fs {
		s = f.MaskString(s)
	}
	return s
}

// ShouldMask .
func (fs Filters) ShouldMask(fieldName string, value interface{}, tag string) bool {
	for _, f := range fs {
		if f.ShouldMask(fieldName, value, tag) {
			return true
		}
	}
	return false
}

func checkShouldMask(fs Filters, fieldName string, value interface{}, tag string) (Filter, bool) {
	for _, f := range fs {
		if f.ShouldMask(fieldName, value, tag) {
			return f, true
		}
	}
	return nil, false
}
