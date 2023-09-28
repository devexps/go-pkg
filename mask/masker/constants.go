package masker

type MType string

const (
	MSecret    MType = "secret"
	MName      MType = "name"
	MPassword  MType = "password"
	MAddress   MType = "address"
	MEmail     MType = "email"
	MMobile    MType = "mobile"
	MTelephone MType = "telephone"
	MURL       MType = "URL"
)

type MaskingCharacter string

const (
	PStar  MaskingCharacter = "*"
	PCross MaskingCharacter = "x"
)
