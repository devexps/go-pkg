package masker

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	MTest   = MType("test")
	StrFunc = func(t MType, s string) (bool, string) {
		if t == MTest {
			return true, s + "-test"
		}
		return false, ""
	}
)

func TestMasker_MaskerOption(t *testing.T) {
	customMasker := NewMasker(WithMaskingCharacter("#"), WithFilteredLabel("[removed]"))

	s := customMasker.String(MPassword, "123456")
	assert.Equal(t, "############", s)

	s1 := customMasker.String(MSecret, "abcdefg123456")
	assert.Equal(t, "[removed]", s1)

	assert.Equal(t, 10, len(customMasker.MarkTypes()))
}

func TestMasker_String(t *testing.T) {
	type args struct {
		t MType
		s string
	}
	defaultMasker := NewMasker()
	customMasker := NewMasker(WithMarkTypes(MName, MTest), WithStringFunc(StrFunc))

	tests := []struct {
		name   string
		masker *Masker
		args   args
		want   string
	}{
		{
			name:   "Error Mask Type",
			masker: defaultMasker,
			args: args{
				t: MType(""),
				s: "abcdefghi",
			},
			want: DefaultFilteredLabel,
		},
		{
			name:   "ID",
			masker: defaultMasker,
			args: args{
				t: MID,
				s: "ABC123456789",
			},
			want: "ABC123****",
		},
		{
			name:   "Name",
			masker: defaultMasker,
			args: args{
				t: MName,
				s: "abcdefghi",
			},
			want: "a**defghi",
		},
		{
			name:   "Password",
			masker: defaultMasker,
			args: args{
				t: MPassword,
				s: "abcdefghi",
			},
			want: "************",
		},
		{
			name:   "Address",
			masker: defaultMasker,
			args: args{
				t: MAddress,
				s: "abcdefghi",
			},
			want: "abcdef******",
		},
		{
			name:   "Email",
			masker: defaultMasker,
			args: args{
				t: MEmail,
				s: "abcd.company@gmail.com",
			},
			want: "abc****mpany@gmail.com",
		},
		{
			name:   "Mobile",
			masker: defaultMasker,
			args: args{
				t: MMobile,
				s: "0987654321",
			},
			want: "0987***321",
		},
		{
			name:   "Telephone",
			masker: defaultMasker,
			args: args{
				t: MTelephone,
				s: "0287654321",
			},
			want: "(02)8765-****",
		},
		{
			name:   "URL",
			masker: defaultMasker,
			args: args{
				t: MURL,
				s: "http://admin:mysecretpassword@localhost:1234/uri",
			},
			want: "http://admin:xxxxx@localhost:1234/uri",
		},
		{
			name:   "CreditCard",
			masker: defaultMasker,
			args: args{
				t: MCreditCard,
				s: "1234567890123456",
			},
			want: "123456******3456",
		},
		{
			name:   "Custom MTest",
			masker: customMasker,
			args: args{
				t: MTest,
				s: "abc",
			},
			want: "abc-test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.masker.String(tt.args.t, tt.args.s); got != tt.want {
				t.Errorf("Masker.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMasker_overlay(t *testing.T) {
	type args struct {
		str     string
		overlay string
		start   int
		end     int
	}
	defaultMasker := NewMasker()

	tests := []struct {
		name          string
		masker        *Masker
		args          args
		wantOverlayed string
	}{
		{
			name:   "Empty Input",
			masker: defaultMasker,
			args: args{
				str:     "",
				overlay: "*",
				start:   0,
				end:     0,
			},
			wantOverlayed: "",
		},
		{
			name:   "Happy Pass",
			masker: defaultMasker,
			args: args{
				str:     "abcdefg",
				overlay: "***",
				start:   1,
				end:     5,
			},
			wantOverlayed: "a***fg",
		},
		{
			name:   "Start Less Than 0",
			masker: defaultMasker,
			args: args{
				str:     "abcdefg",
				overlay: "***",
				start:   -1,
				end:     5,
			},
			wantOverlayed: "***fg",
		},
		{
			name:   "Start Greater Than Length",
			masker: defaultMasker,
			args: args{
				str:     "abcdefg",
				overlay: "***",
				start:   30,
				end:     31,
			},
			wantOverlayed: "abcdefg***",
		},
		{
			name:   "End Less Than 0",
			masker: defaultMasker,
			args: args{
				str:     "abcdefg",
				overlay: "***",
				start:   1,
				end:     -5,
			},
			wantOverlayed: "***bcdefg",
		},
		{
			name:   "Start Less Than End",
			masker: defaultMasker,
			args: args{
				str:     "abcdefg",
				overlay: "***",
				start:   5,
				end:     1,
			},
			wantOverlayed: "a***fg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.masker.overlay(tt.args.str, tt.args.overlay, tt.args.start, tt.args.end); got != tt.wantOverlayed {
				t.Errorf("Masker.overlay() = %v, want %v", got, tt.wantOverlayed)
			}
		})
	}
}

func TestID(t *testing.T) {
	type args struct {
		i string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Empty Input",
			args: args{
				i: "",
			},
			want: "",
		},
		{
			name: "Happy Pass",
			args: args{
				i: "A123456789",
			},
			want: "A12345****",
		},
		{
			name: "Length Less Than 6",
			args: args{
				i: "A12",
			},
			want: "A12****",
		},
		{
			name: "Length Less Than 6",
			args: args{
				i: "A",
			},
			want: "A****",
		},
		{
			name: "Length Between 6 and 10",
			args: args{
				i: "A123456",
			},
			want: "A12345****",
		},
	}
	for _, tt := range tests {
		m := NewMasker()
		t.Run(tt.name, func(t *testing.T) {
			if got := m.ID(tt.args.i); got != tt.want {
				t.Errorf("ID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestName(t *testing.T) {
	type args struct {
		i string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Empty Input",
			args: args{
				i: "",
			},
			want: "",
		},
		{
			name: "Length 1",
			args: args{
				i: "A",
			},
			want: "**",
		},
		{
			name: "Length 2",
			args: args{
				i: "AB",
			},
			want: "A**",
		},
		{
			name: "Length 3",
			args: args{
				i: "ABC",
			},
			want: "A**C",
		},
		{
			name: "Length 4",
			args: args{
				i: "ABCD",
			},
			want: "A**D",
		},
		{
			name: "Length 5",
			args: args{
				i: "ABCDE",
			},
			want: "A**DE",
		},
		{
			name: "Length 6",
			args: args{
				i: "ABCDEF",
			},
			want: "A**DEF",
		},
		{
			name: "English Full Name",
			args: args{
				i: "Jorge Marry",
			},
			want: "J**ge M**ry",
		},
	}
	for _, tt := range tests {
		m := NewMasker()
		t.Run(tt.name, func(t *testing.T) {
			if got := m.Name(tt.args.i); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPassword(t *testing.T) {
	type args struct {
		i string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Empty Input",
			args: args{
				i: "",
			},
			want: "",
		},
		{
			name: "Happy Pass",
			args: args{
				i: "1234567",
			},
			want: "************",
		},
		{
			name: "Happy Pass",
			args: args{
				i: "abcd!@#$%321",
			},
			want: "************",
		},
	}
	for _, tt := range tests {
		m := NewMasker()
		t.Run(tt.name, func(t *testing.T) {
			if got := m.Password(tt.args.i); got != tt.want {
				t.Errorf("Password() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddress(t *testing.T) {
	type args struct {
		i string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Empty Input",
			args: args{
				i: "",
			},
			want: "",
		},
		{
			name: "Long Address",
			args: args{
				i: "1 AB Paradise Road",
			},
			want: "1 AB P******",
		},
		{
			name: "Length Less Than 6",
			args: args{
				i: "ABC",
			},
			want: "******",
		},
	}
	for _, tt := range tests {
		m := NewMasker()
		t.Run(tt.name, func(t *testing.T) {
			if got := m.Address(tt.args.i); got != tt.want {
				t.Errorf("Address() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEmail(t *testing.T) {
	type args struct {
		i string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Empty Input",
			args: args{
				i: "",
			},
			want: "",
		},
		{
			name: "Empty @",
			args: args{
				i: "ggw.changgmail.com",
			},
			want: "ggw****nggmail.com",
		},
		{
			name: "Happy Pass",
			args: args{
				i: "ggw.chang@gmail.com",
			},
			want: "ggw****ng@gmail.com",
		},
		{
			name: "Address Less Than 3",
			args: args{
				i: "qq@gmail.com",
			},
			want: "qq****@gmail.com",
		},
	}
	for _, tt := range tests {
		m := NewMasker()
		t.Run(tt.name, func(t *testing.T) {
			if got := m.Email(tt.args.i); got != tt.want {
				t.Errorf("Email() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMobile(t *testing.T) {
	type args struct {
		i string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Empty Input",
			args: args{
				i: "",
			},
			want: "",
		},
		{
			name: "Happy Pass",
			args: args{
				i: "0978978978",
			},
			want: "0978***978",
		},
		{
			name: "Happy Pass",
			args: args{
				i: "0912345678",
			},
			want: "0912***678",
		},
	}
	for _, tt := range tests {
		m := NewMasker()
		t.Run(tt.name, func(t *testing.T) {
			if got := m.Mobile(tt.args.i); got != tt.want {
				t.Errorf("Mobile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTelephone(t *testing.T) {
	type args struct {
		i string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Empty Input",
			args: args{
				i: "",
			},
			want: "",
		},
		{
			name: "Out of range",
			args: args{
				i: "(02-)27-99-3--078-4325",
			},
			want: "02279930784325",
		},
		{
			name: "With Special Chart",
			args: args{
				i: "(02-)27   99-3--078",
			},
			want: "(02)2799-****",
		},
		{
			name: "Happy Pass",
			args: args{
				i: "0227993078",
			},
			want: "(02)2799-****",
		},
		{
			name: "Happy Pass",
			args: args{
				i: "0788079966",
			},
			want: "(07)8807-****",
		},
	}
	for _, tt := range tests {
		m := NewMasker()
		t.Run(tt.name, func(t *testing.T) {
			if got := m.Telephone(tt.args.i); got != tt.want {
				t.Errorf("Telephone() = %v, want %v", got, tt.want)
			}
		})
	}
}
