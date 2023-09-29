package masker

import (
	"math"
	"net/url"
	"strings"
)

type StringFunc func(t MType, s string) (bool, string)

type Option func(o *Masker)

type Masker struct {
	maskingCharacter string
	filteredLabel    string
	maskTypes        []MType
	stringFunc       StringFunc
}

// NewMasker create a Masker instance
func NewMasker(opts ...Option) *Masker {
	masker := &Masker{
		maskingCharacter: string(PStar),
		filteredLabel:    DefaultFilteredLabel,
		maskTypes:        []MType{MSecret, MID, MName, MPassword, MAddress, MEmail, MMobile, MTelephone, MURL, MCreditCard},
	}
	for _, opt := range opts {
		opt(masker)
	}
	return masker
}

// WithMaskingCharacter sets the custom masking character
func WithMaskingCharacter(mask string) Option {
	return func(o *Masker) {
		o.maskingCharacter = mask
	}
}

// WithFilteredLabel sets the custom filtered label
func WithFilteredLabel(label string) Option {
	return func(o *Masker) {
		o.filteredLabel = label
	}
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
func (m *Masker) String(t MType, s string) string {
	if m.stringFunc != nil {
		if b, o := m.stringFunc(t, s); b {
			return o
		}
	}
	switch t {
	case MID:
		return m.ID(s)
	case MName:
		return m.Name(s)
	case MPassword:
		return m.Password(s)
	case MAddress:
		return m.Address(s)
	case MEmail:
		return m.Email(s)
	case MMobile:
		return m.Mobile(s)
	case MTelephone:
		return m.Telephone(s)
	case MURL:
		return m.URL(s)
	case MCreditCard:
		return m.CreditCard(s)
	default:
		return m.filteredLabel
	}
}

// ID mask last 4 digits of ID number
//
// Example:
//   input: ABC123456789
//   output: ABC12345****
func (m *Masker) ID(i string) string {
	l := len([]rune(i))
	if l == 0 {
		return ""
	}
	return m.overlay(i, strLoop(m.maskingCharacter, len("****")), 6, l)
}

// Name mask the second letter and the third letter
//
// Example:
//   input: ABCD
//   output: A**D
func (m *Masker) Name(s string) string {
	l := len([]rune(s))
	if l == 0 {
		return ""
	}
	// if it has space
	if strs := strings.Split(s, " "); len(strs) > 1 {
		tmp := make([]string, len(strs))
		for idx, str := range strs {
			tmp[idx] = m.Name(str)
		}
		return strings.Join(tmp, " ")
	}
	if l == 2 || l == 3 {
		return m.overlay(s, strLoop(m.maskingCharacter, len("**")), 1, 2)
	}
	if l > 3 {
		return m.overlay(s, strLoop(m.maskingCharacter, len("**")), 1, 3)
	}
	return strLoop(m.maskingCharacter, len("**"))
}

// Password always return "************"
func (m *Masker) Password(s string) string {
	l := len([]rune(s))
	if l == 0 {
		return ""
	}
	return strLoop(m.maskingCharacter, len("************"))
}

// Address keep first 6 letters, mask the rest
//
// Example:
//   input: Cecilia Chapman 711-2880 Nulla St. Mankato Mississippi 96522
//   output: Cecili******
func (m *Masker) Address(s string) string {
	l := len([]rune(s))
	if l == 0 {
		return ""
	}
	n := 6
	if l <= n {
		return strLoop(m.maskingCharacter, len("******"))
	}
	return m.overlay(s, strLoop(m.maskingCharacter, len("******")), n, math.MaxInt64)
}

// Email keep domain and the first 3 letters
//
// Example:
//   input: abcd.company@gmail.com
//   output: abc****@gmail.com
func (m *Masker) Email(s string) string {
	l := len([]rune(s))
	if l == 0 {
		return ""
	}
	tmp := strings.Split(s, "@")

	switch len(tmp) {
	case 0, 1:
		return m.overlay(s, strLoop(m.maskingCharacter, len("****")), 3, 7)
	}
	addr := tmp[0]
	domain := tmp[1]

	addr = m.overlay(addr, strLoop(m.maskingCharacter, len("****")), 3, 7)
	return addr + "@" + domain
}

// Mobile mask 3 digits from the 4'th digit
//
// Example:
//   input: 0987654321
//   output: 0987***321
func (m *Masker) Mobile(s string) string {
	if len(s) == 0 {
		return ""
	}
	return m.overlay(s, strLoop(m.maskingCharacter, len("***")), 4, 7)
}

// Telephone remove "(", ")", " ", "-" chart, and mask last 4 digits of telephone number, format to "(??)????-****"
//
// Example:
//   input: 0287654321
//   output: (02)8765-****"
func (m *Masker) Telephone(s string) string {
	l := len([]rune(s))
	if l == 0 {
		return ""
	}
	s = strings.Replace(s, " ", "", -1)
	s = strings.Replace(s, "(", "", -1)
	s = strings.Replace(s, ")", "", -1)
	s = strings.Replace(s, "-", "", -1)

	l = len([]rune(s))

	if l != 10 && l != 8 {
		return s
	}
	ans := ""

	if l == 10 {
		ans += "("
		ans += s[:2]
		ans += ")"
		s = s[2:]
	}
	ans += s[:4]
	ans += "-"
	ans += "****"

	return ans
}

// URL mask the password part of the URL if exists
//
// Example:
//   input: http://admin:mysecretpassword@localhost:1234/uri
//   output:http://admin:xxxxx@localhost:1234/uri
func (m *Masker) URL(s string) string {
	u, err := url.Parse(s)
	if err != nil {
		return s
	}
	return u.Redacted()
}

// CreditCard mask 6 digits from the 7'th digit
//
// Example:
//   input1: 1234567890123456 (VISA, JCB, MasterCard)(len = 16)
//   output1: 123456******3456
//   input2: 123456789012345` (American Express)(len = 15)
//   output2: 123456******345`
func (m *Masker) CreditCard(i string) string {
	l := len([]rune(i))
	if l == 0 {
		return ""
	}
	return m.overlay(i, strLoop(m.maskingCharacter, len("******")), 6, 12)
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
