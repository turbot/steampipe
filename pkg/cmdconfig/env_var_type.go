package cmdconfig

type EnvVarType int

const (
	String EnvVarType = iota
	Int
	Bool
)

func (t EnvVarType) String() string {
	switch t {
	case String:
		return "string"
	case Bool:
		return "boolean"
	case Int:
		return "int"
	}
	return "unknown"
}

//go:generate go run golang.org/x/tools/cmd/stringer -type=EnvVarType
