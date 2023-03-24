package cmdconfig

type EnvVarType int

const (
	String EnvVarType = iota
	Int
	Bool
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=EnvVarType
