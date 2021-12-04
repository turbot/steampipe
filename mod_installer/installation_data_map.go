package mod_installer

import (
	"fmt"
	"strings"
)

type InstallationDataMap map[string]*InstallationData

func (i InstallationDataMap) InstallReport() string {
	if len(i) == 0 {
		return "No dependencies installed"
	}
	strs := make([]string, len(i))
	idx := 0
	for name, d := range i {
		strs[idx] = fmt.Sprintf("%s@%s", name, d.Version.String())
		idx++
	}
	return fmt.Sprintf("\nInstalled %d dependencies:\n  - %s\n", len(i), strings.Join(strs, "\n  - "))
}

func (i InstallationDataMap) SumFileText() string {
	var strs = make([]string, len(i))
	idx := 0
	for _, d := range i {
		strs[idx] = fmt.Sprintf("%s v%s", d.Name, d.Version.String())
		idx++
	}
	return strings.Join(strs, "\n")
}
