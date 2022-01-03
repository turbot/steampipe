package modinstaller

import (
	"fmt"
	"testing"

	"github.com/Masterminds/semver"
)

func TestModInstaller(t *testing.T) {
	cs, err := semver.NewConstraint("^3")
	v, _ := semver.NewVersion("3.1")
	res := cs.Check(v)
	fmt.Println(res)

	fmt.Println(cs)
	fmt.Println(err)
}
