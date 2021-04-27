package parse

type LoadModFlag uint32

const (
	CreateDefaultMod LoadModFlag = 1 << iota
	CreatePseudoResources
)

type ParseModOptions struct {
	Flags   LoadModFlag
	Exclude []string
}

func (o *ParseModOptions) CreateDefaultMod() bool {
	return o.Flags&CreateDefaultMod == CreateDefaultMod
}

func (o *ParseModOptions) CreatePseudoResources() bool {
	return o.Flags&CreatePseudoResources == CreatePseudoResources
}
