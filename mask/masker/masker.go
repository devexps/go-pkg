package masker

import (
	"math"
	"net/url"
	"strings"
)

type StringFunc func(t MType, i string) (bool, string)

type Option func(o *Masker)

type Masker struct {
	mask       string
	maskTypes  []MType
	stringFunc StringFunc
}

// NewMasker create a Masker instance
func NewMasker(opts ...Option) *Masker {
	masker := &Masker{
		mask:      string(PStar),
		maskTypes: []MType{MSecret, MName, MPassword, MAddress, MEmail, MMobile, MTelephone, MURL},
	}
	for _, opt := range opts {
		opt(masker)
	}
	return masker
}

// WithMarkTypes adds your custom mark types
func WithMarkTypes(markType ...MType) Option {
	return func(o *Masker) {
		for _, nt := range markType {
			for _, ot := range o.maskTypes {
				if ot == nt {
					continue
				}
			}
			o.maskTypes = append(o.maskTypes, nt)
		}
	}
}

// WithStringFunc sets the stringFunc
func WithStringFunc(f StringFunc) Option {
	return func(o *Masker) {
		o.stringFunc = f
	}
}

// MarkTypes returns available mark types
func (m *Masker) MarkTypes() []MType {
	return m.maskTypes
}

// String mask input string of the mask type
func (m *Masker) String(t MType, i string, defaultFiltered string) string {
	if m.stringFunc != nil {
		if b, o := m.stringFunc(t, i); b {
			return o
		}
	}
	switch t {
	case MName:
		return m.Name(i)
	case MPassword:
		return m.Password(i)
	case MAddress:
		return m.Address(i)
	case MEmail:
		return m.Email(i)
	case MMobile:
		return m.Mobile(i)
	case MTelephone:
		return m.Telephone(i)
	case MURL:
		return m.URL(i)
	default:
		return defaultFiltered
	}
}

// Name mask the second letter and the third letter
//
// Example:
//   input: ABCD
//   output: A**D
func (m *Masker) Name(i string) string {
	l := len([]rune(i))
	if l == 0 {
		return ""
	}
	// if it has space
	if strs := strings.Split(i, " "); len(strs) > 1 {
		tmp := make([]string, len(strs))
		for idx, str := range strs {
			tmp[idx] = m.Name(str)
		}
		return strings.Join(tmp, " ")
	}
	if l == 2 || l == 3 {
		return m.overlay(i, strLoop(m.mask, len("**")), 1, 2)
	}
	if l > 3 {
		return m.overlay(i, strLoop(m.mask, len("**")), 1, 3)
	}
	return strLoop(m.mask, len("**"))
}

// Password always return "************"
func (m *Masker) Password(i string) string {
	l := len([]rune(i))
	if l == 0 {
		return ""
	}
	return strLoop(m.mask, len("************"))
}

// Address keep first 6 letters, mask the rest
//
// Example:
//   input: Cecilia Chapman 711-2880 Nulla St. Mankato Mississippi 96522
//   output: Cecili******
func (m *Masker) Address(i string) string {
	l := len([]rune(i))
	if l == 0 {
		return ""
	}
	n := 6
	if l <= n {
		return strLoop(m.mask, len("******"))
	}
	return m.overlay(i, strLoop(m.mask, len("******")), n, math.MaxInt64)
}

// Email keep domain and the first 3 letters
//
// Example:
//   input: abcd.company@gmail.com
//   output: abc****@gmail.com
func (m *Masker) Email(i string) string {
	l := len([]rune(i))
	if l == 0 {
		return ""
	}
	tmp := strings.Split(i, "@")

	switch len(tmp) {
	case 0:
		return ""
	case 1:
		return m.overlay(i, strLoop(m.mask, len("****")), 3, len(tmp[0]))
	}
	addr := tmp[0]
	domain := tmp[1]

	addr = m.overlay(addr, strLoop(m.mask, len("****")), 3, len(tmp[0]))
	return addr + "@" + domain
}

// Mobile mask 3 digits from the 4'th digit
//
// Example:
//   input: 0987654321
//   output: 0987***321
func (m *Masker) Mobile(i string) string {
	if len(i) == 0 {
		return ""
	}
	return m.overlay(i, strLoop(m.mask, len("***")), 4, 7)
}

// Telephone remove "(", ")", " ", "-" chart, and mask last 4 digits of telephone number, format to "(??)????-****"
//
// Example:
//   input: 0287654321
//   output: (02)8765-****"
func (m *Masker) Telephone(i string) string {
	l := len([]rune(i))
	if l == 0 {
		return ""
	}
	i = strings.Replace(i, " ", "", -1)
	i = strings.Replace(i, "(", "", -1)
	i = strings.Replace(i, ")", "", -1)
	i = strings.Replace(i, "-", "", -1)

	l = len([]rune(i))

	if l != 10 && l != 8 {
		return i
	}
	ans := ""

	if l == 10 {
		ans += "("
		ans += i[:2]
		ans += ")"
		i = i[2:]
	}
	ans += i[:4]
	ans += "-"
	ans += "****"

	return ans
}

// URL mask the password part of the URL if exists
//
// Example:
//   input: http://admin:mysecretpassword@localhost:1234/uri
//   output:http://admin:xxxxx@localhost:1234/uri
func (m *Masker) URL(i string) string {
	u, err := url.Parse(i)
	if err != nil {
		return i
	}
	return u.Redacted()
}

func (m *Masker) overlay(str string, overlay string, start int, end int) (overlayed string) {
	r := []rune(str)
	l := len([]rune(r))
	if l == 0 {
		return ""
	}
	if start < 0 {
		start = 0
	}
	if start > l {
		start = l
	}
	if end < 0 {
		end = 0
	}
	if end > l {
		end = l
	}
	if start > end {
		tmp := start
		start = end
		end = tmp
	}
	overlayed = ""
	overlayed += string(r[:start])
	overlayed += overlay
	overlayed += string(r[end:])
	return overlayed
}

func strLoop(str string, length int) string {
	var mask string
	for i := 1; i <= length; i++ {
		mask += str
	}
	return mask
}
