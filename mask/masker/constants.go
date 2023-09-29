package masker

type MType string

const (
	MSecret     MType = "secret"
	MID         MType = "id"
	MName       MType = "name"
	MPassword   MType = "password"
	MAddress    MType = "address"
	MEmail      MType = "email"
	MMobile     MType = "mobile"
	MTelephone  MType = "telephone"
	MURL        MType = "url"
	MCreditCard MType = "credit"

	DefaultFilteredLabel = "[filtered]"
)

type MaskingCharacter string

const (
	PStar  MaskingCharacter = "*"
	PCross MaskingCharacter = "x"
)
