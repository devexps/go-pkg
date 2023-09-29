package mask

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/devexps/go-pkg/v2/mask/masker"
)

func TestTagFilter(t *testing.T) {
	t.Run("default ", func(t *testing.T) {
		type myRecord struct {
			ID    string
			EMail *string `mask:"secret"`
		}
		e := "dummy@dummy.com"
		record := myRecord{
			ID:    "userId",
			EMail: &e,
		}
		maskTool := New(TagFilter())
		filteredData := maskTool.Apply(record)
		assert.NotNil(t, filteredData)
		copied, ok := filteredData.(myRecord)
		require.True(t, ok)
		require.NotNil(t, copied)
		assert.Equal(t, masker.DefaultFilteredLabel, *copied.EMail)
		assert.Equal(t, record.ID, copied.ID)
	})

	t.Run("custom ", func(t *testing.T) {
		type myRecord struct {
			ID    string
			EMail string `mask:"email"`
			Token string `mask:"secret"`
		}
		record := myRecord{
			ID:    "userId",
			EMail: "dummy@dummy.com",
			Token: "abcd1234",
		}
		maskTool := New(TagsFilter())
		filteredData := maskTool.Apply(record)
		assert.NotNil(t, filteredData)
		copied, ok := filteredData.(myRecord)
		require.True(t, ok)
		require.NotNil(t, copied)
		assert.Equal(t, "dum****@dummy.com", copied.EMail)
		assert.Equal(t, masker.DefaultFilteredLabel, copied.Token)
		assert.Equal(t, record.ID, copied.ID)
	})
}

func TestFieldFilter(t *testing.T) {
	t.Run("default", func(*testing.T) {
		type myRecord struct {
			ID    string
			Phone string
			Email string
		}
		record := myRecord{
			ID:    "userId",
			Phone: "090-0000-0000",
			Email: "abc@gmail.com",
		}
		maskTool := New(FieldFilter("Phone"), FieldFilter("Email", masker.MEmail))
		filteredData := maskTool.Apply(record)
		require.NotNil(t, filteredData)
		copied, ok := filteredData.(myRecord)
		require.True(t, ok)
		require.NotNil(t, copied)
		assert.Equal(t, masker.DefaultFilteredLabel, copied.Phone)
		assert.Equal(t, "abc****@gmail.com", copied.Email)
		assert.Equal(t, record.ID, copied.ID)
	})
}

func TestFieldPrefixFilter(t *testing.T) {
	t.Run("default", func(*testing.T) {
		type myRecord struct {
			ID          string
			SecurePhone string
			SecureEmail string
		}
		record := myRecord{
			ID:          "userId",
			SecurePhone: "090-0000-0000",
			SecureEmail: "abc.def@gmail.com",
		}
		maskTool := New(FieldPrefixFilter("Secure", masker.MMobile))
		filteredData := maskTool.Apply(record)
		require.NotNil(t, filteredData)
		copied, ok := filteredData.(myRecord)
		require.True(t, ok)
		require.NotNil(t, copied)
		assert.Equal(t, "090-***0-0000", copied.SecurePhone)
		assert.Equal(t, "abc.***@gmail.com", copied.SecureEmail)
		assert.Equal(t, record.ID, copied.ID)
	})
}

func TestFieldRegexFilter(t *testing.T) {
	type myRecord struct {
		ID    string
		Link  string
		Link1 string
	}
	customRegex := "^https:\\/\\/(dummy-backend.)[0-9a-z]*.com\\b([-a-zA-Z0-9@:%_\\+.~#?&//=]*)$"

	t.Run("default", func(*testing.T) {
		record := myRecord{
			ID:    "userId",
			Link:  "https://dummy-backend.dummy.com/v2/random",
			Link1: "https://dummy-frontend.dummy.com/v2/random",
		}
		maskTool := New(RegexFilter(customRegex))
		filteredData := maskTool.Apply(record)
		require.NotNil(t, filteredData)
		copied, ok := filteredData.(myRecord)
		require.True(t, ok)
		require.NotNil(t, copied)
		assert.Equal(t, masker.DefaultFilteredLabel, copied.Link)
		assert.Equal(t, record.Link1, copied.Link1)
		assert.Equal(t, record.ID, copied.ID)
	})

	t.Run("masking", func(*testing.T) {
		record := myRecord{
			ID:    "userId",
			Link:  "https://dummy-backend.dummy.com/v2/random",
			Link1: "https://dummy-frontend.dummy.com/v2/random",
		}
		maskTool := New(RegexFilter(customRegex, masker.MPassword))
		filteredData := maskTool.Apply(record)
		require.NotNil(t, filteredData)
		copied, ok := filteredData.(myRecord)
		require.True(t, ok)
		require.NotNil(t, copied)
		assert.Equal(t, "************", copied.Link)
		assert.Equal(t, record.Link1, copied.Link1)
		assert.Equal(t, record.ID, copied.ID)
	})
}

func TestValueFilter(t *testing.T) {
	t.Run("DefaultValueFilter", func(t *testing.T) {
		const issuedToken = "abcd1234"
		maskTool := New(ValueFilter(issuedToken))

		t.Run("string", func(t *testing.T) {
			record := "Authorization: Bearer " + issuedToken
			filteredData := maskTool.Apply(record)
			require.NotNil(t, filteredData)
			assert.Equal(t, "Authorization: Bearer [filtered]", filteredData)
		})
		t.Run("struct", func(t *testing.T) {
			type myRecord struct {
				ID   string
				Data string
			}
			record := myRecord{
				ID:   "userId",
				Data: issuedToken,
			}

			filteredData := maskTool.Apply(record)
			require.NotNil(t, filteredData)
			copied, ok := filteredData.(myRecord)
			require.True(t, ok)
			require.NotNil(t, copied)
			assert.Equal(t, record.ID, copied.ID)
			assert.Equal(t, masker.DefaultFilteredLabel, copied.Data)
		})
		t.Run("array", func(t *testing.T) {
			record := []string{
				"userId",
				"data",
				issuedToken,
			}
			filteredData := maskTool.Apply(record)
			require.NotNil(t, filteredData)
			assert.Equal(t, []string([]string{"userId", "data", masker.DefaultFilteredLabel}), filteredData)
		})
		t.Run("map", func(*testing.T) {
			mapRecord := map[string]interface{}{
				"data": issuedToken,
			}
			filteredData := maskTool.Apply(mapRecord)
			require.NotNil(t, filteredData)
			assert.Equal(t, map[string]interface{}{"data": "[filtered]"}, filteredData)
		})
	})
	t.Run("CustomValueFilter", func(t *testing.T) {
		const issuedToken = "abcd1234"
		maskTool := New(ValueFilter(issuedToken, masker.MPassword))
		t.Run("string", func(t *testing.T) {
			record := "Authorization: Bearer " + issuedToken
			filteredData := maskTool.Apply(record)
			require.NotNil(t, filteredData)
			assert.Equal(t, "Authorization: Bearer ************", filteredData)
		})
	})
	t.Run("OtherValueFilter", func(t *testing.T) {
		customMasker := New(
			ValueFilter("blue"),
		)
		type testData struct {
			ID    int
			Name  string
			Label string
		}
		t.Run("non-ptr struct can be modified", func(t *testing.T) {
			data := testData{
				Name:  "blue",
				Label: "five",
			}
			v := customMasker.Apply(data)
			require.NotNil(t, v)
			copied, ok := v.(testData)
			require.True(t, ok)
			require.NotNil(t, copied)
			assert.Equal(t, masker.DefaultFilteredLabel, copied.Name)
			assert.Equal(t, "five", copied.Label)
		})
		t.Run("original data is not modified when filtered", func(t *testing.T) {
			data := &testData{
				ID:    100,
				Name:  "blue",
				Label: "five",
			}
			v := customMasker.Apply(data)
			require.NotNil(t, v)
			copied, ok := v.(*testData)
			require.True(t, ok)
			require.NotNil(t, copied)
			assert.Equal(t, masker.DefaultFilteredLabel, copied.Name)
			assert.Equal(t, "blue", data.Name)
			assert.Equal(t, "five", data.Label)
			assert.Equal(t, "five", copied.Label)
			assert.Equal(t, 100, copied.ID)
		})
		t.Run("nested structure can be modified", func(t *testing.T) {
			type testDataParent struct {
				Child testData
			}
			data := &testDataParent{
				Child: testData{
					Name:  "blue",
					Label: "five",
				},
			}
			v := customMasker.Apply(data)
			require.NotNil(t, v)
			copied, ok := v.(*testDataParent)
			require.True(t, ok)
			require.NotNil(t, copied)
			assert.Equal(t, masker.DefaultFilteredLabel, copied.Child.Name)
			assert.Equal(t, "five", copied.Child.Label)
		})
		t.Run("map data", func(t *testing.T) {
			data := map[string]*testData{
				"xyz": {
					Name:  "blue",
					Label: "five",
				},
			}
			v := customMasker.Apply(data)
			require.NotNil(t, v)
			copied, ok := v.(map[string]*testData)
			require.True(t, ok)
			require.NotNil(t, copied)
			assert.Equal(t, masker.DefaultFilteredLabel, copied["xyz"].Name)
			assert.Equal(t, "five", copied["xyz"].Label)
		})
		t.Run("array data with ptr", func(t *testing.T) {
			data := []*testData{
				{
					Name:  "orange",
					Label: "five",
				},
				{
					Name:  "blue",
					Label: "five",
				},
			}
			v := customMasker.Apply(data)
			require.NotNil(t, v)
			copied, ok := v.([]*testData)
			require.True(t, ok)
			require.NotNil(t, copied)
			assert.Equal(t, "orange", copied[0].Name)
			assert.Equal(t, masker.DefaultFilteredLabel, copied[1].Name)
			assert.Equal(t, "five", copied[1].Label)
		})
	})
}

func TestTypeFilter(t *testing.T) {
	type password string
	type array []int32

	t.Run("CustomTypeFilter", func(t *testing.T) {
		type myRecord struct {
			ID       string
			Password password
		}
		record := myRecord{
			ID:       "userId",
			Password: "abcd1234",
		}
		t.Run("Type Filter with Mask Type", func(t *testing.T) {
			maskTool := New(TypeFilter(password(""), masker.MPassword))
			filteredData := maskTool.Apply(record)
			require.NotNil(t, filteredData)
			copied, ok := filteredData.(myRecord)
			require.True(t, ok)
			require.NotNil(t, copied)
			assert.Equal(t, password("************"), copied.Password)
			assert.Equal(t, record.ID, copied.ID)

		})
	})
	t.Run("TypeFilter", func(t *testing.T) {
		type myRecord struct {
			ID       string
			Password password
			Ages     array
		}
		record := myRecord{
			ID:       "userId",
			Password: "abcd1234",
			Ages:     array{10},
		}
		t.Run("Default Type Filter", func(t *testing.T) {
			maskTool := New(TypeFilter(password("")), TypeFilter(array{}))
			filteredData := maskTool.Apply(record)
			require.NotNil(t, filteredData)
			copied, ok := filteredData.(myRecord)
			require.True(t, ok)
			require.NotNil(t, copied)
			assert.Equal(t, password(masker.DefaultFilteredLabel), copied.Password)
			assert.Equal(t, array(nil), copied.Ages)
			assert.Equal(t, record.ID, copied.ID)
		})
	})
}

func TestCustomMasker(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		SetMasker(masker.NewMasker(masker.WithMarkTypes("test")))
		assert.NotNil(t, maskerInstance)

		maskTool := New()
		assert.NotNil(t, maskTool)
		filteredData := maskTool.Apply(nil)
		assert.Nil(t, filteredData)
	})
}
