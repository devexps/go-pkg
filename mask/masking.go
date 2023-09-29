package mask

import (
	"reflect"

	"github.com/devexps/go-pkg/v2/mask/masker"
)

var (
	maskerInstance = masker.NewMasker()
)

// SetMasker sets a custom masker instance
func SetMasker(masker *masker.Masker) {
	maskerInstance = masker
}

type Masking interface {
	Apply(v interface{}) interface{}
}

type masking struct {
	filters Filters
}

// New creates a new Masking instance
func New(filters ...Filter) Masking {
	return &masking{
		filters: append(Filters{}, filters...),
	}
}

// Apply returns a copy interface with masked fields
func (m *masking) Apply(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	return m.clone("", reflect.ValueOf(v), "").Interface()
}

func (m *masking) clone(fieldName string, value reflect.Value, tag string) reflect.Value {
	adjustValue := func(ret reflect.Value) reflect.Value {
		switch value.Kind() {
		case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Array:
			return ret
		default:
			return ret.Elem()
		}
	}
	src := value

	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return reflect.New(value.Type()).Elem()
		}
		src = value.Elem()
	}
	var dst reflect.Value

	maskingFilter, shouldMask := checkShouldMask(m.filters, fieldName, src.Interface(), tag)
	if shouldMask {
		dst = reflect.New(src.Type())
		switch src.Kind() {
		case reflect.String:
			filteredData := maskingFilter.MaskString(value.String())
			dst.Elem().SetString(filteredData)
		case reflect.Array, reflect.Slice:
			dst = dst.Elem()
		}
		return adjustValue(dst)
	}
	switch src.Kind() {
	case reflect.String:
		dst = reflect.New(src.Type())
		filtered := m.filters.ReplaceString(value.String())
		dst.Elem().SetString(filtered)
	case reflect.Struct:
		dst = reflect.New(src.Type())
		t := src.Type()

		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			fv := src.Field(i)
			if !fv.CanInterface() {
				continue
			}
			tagValue := f.Tag.Get(tagKey)
			if fv.Type().Kind() == reflect.Ptr && fv.Elem().Kind() == reflect.String {
				a := m.clone(f.Name, fv.Elem(), tagValue).Convert(reflect.TypeOf("")).Interface().(string)
				dst.Elem().Field(i).Set(reflect.New(fv.Elem().Type()))
				dst.Elem().Field(i).Elem().SetString(a)
			} else {
				dst.Elem().Field(i).Set(m.clone(f.Name, fv, tagValue))
			}
		}
	case reflect.Map:
		dst = reflect.MakeMap(src.Type())
		keys := src.MapKeys()
		for i := 0; i < src.Len(); i++ {
			mValue := src.MapIndex(keys[i])
			dst.SetMapIndex(keys[i], m.clone(keys[i].String(), mValue, ""))
		}
	case reflect.Array, reflect.Slice:
		dst = reflect.MakeSlice(src.Type(), src.Len(), src.Cap())
		for i := 0; i < src.Len(); i++ {
			dst.Index(i).Set(m.clone(fieldName, src.Index(i), ""))
		}
	case reflect.Interface:
		dst = reflect.New(src.Type())
		data := value.Interface()
		stringData, ok := data.(string)
		if !ok {
			dst.Elem().Set(src)
		} else {
			filtered := m.filters.ReplaceString(stringData)
			dst.Elem().Set(reflect.ValueOf(filtered))
		}
	default:
		dst = reflect.New(src.Type())
		dst.Elem().Set(src)
	}
	return adjustValue(dst)
}
