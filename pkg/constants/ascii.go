package constants

var (
	// OptLeftArrowASCIICode ::
	OptLeftArrowASCIICode = []byte{0x1b, 0x62}
	// OptRightArrowASCIICode ::
	OptRightArrowASCIICode = []byte{0x1b, 0x66}
	// AltLeftArrowASCIICode ::
	AltLeftArrowASCIICode = []byte{0x1b, 0x1b, 0x5B, 0x44}
	// AltRightArrowASCIICode ::
	AltRightArrowASCIICode = []byte{0x1b, 0x1b, 0x5B, 0x43}
)

var (
	SuppressedASCIICodes = [][]byte{
		// Alt+UpArrow
		{27, 27, 91, 65},
		// Alt+DownArrow
		{27, 27, 91, 66},
	}

	SuppressedASCIICodesForWSL = [][]byte{
		// Alt+D
		{195, 176},
	}
)
