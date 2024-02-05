package modinstaller

import (
	"bytes"
	"fmt"
	"os"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/turbot/steampipe/pkg/steampipeconfig/modconfig"
	"github.com/zclconf/go-cty/cty"
)

// updates the 'require' block in 'mod.sp'
func (i *ModInstaller) updateModFile() error {
	contents, err := i.loadModFileBytes()
	if err != nil {
		return err
	}

	oldRequire := i.oldRequire
	newRequire := i.workspaceMod.Require

	// fill these requires in with empty requires
	// so that we don't have to do nil checks everywhere
	// from here on out - if it's empty - it's nil

	if oldRequire == nil {
		// use an empty require as the old requirements
		oldRequire = modconfig.NewRequire()
	}
	if newRequire == nil {
		// use a stub require instance
		newRequire = modconfig.NewRequire()
	}

	changes := EmptyChangeSet()

	if i.shouldDeleteRequireBlock(oldRequire, newRequire) {
		changes = i.buildChangeSetForRequireDelete(oldRequire, newRequire)
	} else if i.shouldCreateRequireBlock(oldRequire, newRequire) {
		changes = i.buildChangeSetForRequireCreate(oldRequire, newRequire)
	} else if !newRequire.Empty() && !oldRequire.Empty() {
		changes = i.calculateChangeSet(oldRequire, newRequire)
	}

	if len(changes) == 0 {
		// nothing to do here
		return nil
	}

	contents.ApplyChanges(changes)
	contents.Apply(hclwrite.Format)

	return os.WriteFile(i.workspaceMod.FilePath(), contents.Bytes(), 0644)
}

// loads the contents of the mod.sp file and wraps it with a thin wrapper
// to assist in byte sequence manipulation
func (i *ModInstaller) loadModFileBytes() (*ByteSequence, error) {
	modFileBytes, err := os.ReadFile(i.workspaceMod.FilePath())
	if err != nil {
		return nil, err
	}
	return NewByteSequence(modFileBytes), nil
}

func (i *ModInstaller) shouldDeleteRequireBlock(oldRequire *modconfig.Require, newRequire *modconfig.Require) bool {
	return newRequire.Empty() && !oldRequire.Empty()
}

func (i *ModInstaller) shouldCreateRequireBlock(oldRequire *modconfig.Require, newRequire *modconfig.Require) bool {
	return !newRequire.Empty() && oldRequire.Empty()
}

func (i *ModInstaller) buildChangeSetForRequireDelete(oldRequire *modconfig.Require, newRequire *modconfig.Require) ChangeSet {
	return NewChangeSet(&Change{
		Operation:   Delete,
		OffsetStart: oldRequire.DeclRange.Start.Byte,
		OffsetEnd:   oldRequire.BodyRange.End.Byte,
	})
}

func (i *ModInstaller) buildChangeSetForRequireCreate(oldRequire *modconfig.Require, newRequire *modconfig.Require) ChangeSet {
	// if the new require is not empty, but the old one is
	// add a new require block with the new stuff
	// by generating the HCL string that goes in
	f := hclwrite.NewEmptyFile()

	var body *hclwrite.Body
	var insertOffset int

	if oldRequire.BodyRange.Start.Byte != 0 {
		// this means that there is a require block
		// but is probably empty
		body = f.Body()
		insertOffset = oldRequire.BodyRange.End.Byte - 1
	} else {
		// we don't have a require block at all
		// let's create one to append to
		body = f.Body().AppendNewBlock("require", nil).Body()
		insertOffset = i.workspaceMod.DeclRange.End.Byte - 1
	}

	for _, mvc := range newRequire.Mods {
		newBlock := i.createNewModRequireBlock(mvc)
		body.AppendBlock(newBlock)
	}

	// prefix and suffix with new lines
	// this is so that we can handle empty blocks
	// which do not have newlines
	buffer := bytes.NewBuffer([]byte{'\n'})
	buffer.Write(f.Bytes())
	buffer.WriteByte('\n')

	return NewChangeSet(&Change{
		Operation:   Insert,
		OffsetStart: insertOffset,
		Content:     buffer.Bytes(),
	})
}

func (i *ModInstaller) calculateChangeSet(oldRequire *modconfig.Require, newRequire *modconfig.Require) ChangeSet {
	if oldRequire.Empty() && newRequire.Empty() {
		// both are empty
		// nothing to do
		return EmptyChangeSet()
	}
	// calculate the changes
	uninstallChanges := i.calcChangesForUninstall(oldRequire, newRequire)
	installChanges := i.calcChangesForInstall(oldRequire, newRequire)
	updateChanges := i.calcChangesForUpdate(oldRequire, newRequire)

	return MergeChangeSet(
		uninstallChanges,
		installChanges,
		updateChanges,
	)
}

// creates a new "mod" block which can be written as part of the "require" block in mod.sp
func (i *ModInstaller) createNewModRequireBlock(modVersion *modconfig.ModVersionConstraint) *hclwrite.Block {
	modRequireBlock := hclwrite.NewBlock("mod", []string{modVersion.Name})
	modRequireBlock.Body().SetAttributeValue("version", cty.StringVal(modVersion.VersionString))
	return modRequireBlock
}

// calculates changes required in mod.sp to reflect uninstalls
func (i *ModInstaller) calcChangesForUninstall(oldRequire *modconfig.Require, newRequire *modconfig.Require) ChangeSet {
	changes := ChangeSet{}
	for _, requiredMod := range oldRequire.Mods {
		// check if this mod is still a dependency
		if modInNew := newRequire.GetModDependency(requiredMod.Name); modInNew == nil {
			changes = append(changes, &Change{
				Operation:   Delete,
				OffsetStart: requiredMod.DefRange.Start.Byte,
				OffsetEnd:   requiredMod.BodyRange.End.Byte,
			})
		}
	}
	return changes
}

// calculates changes required in mod.sp to reflect new installs
func (i *ModInstaller) calcChangesForInstall(oldRequire *modconfig.Require, newRequire *modconfig.Require) ChangeSet {
	modsToAdd := []*modconfig.ModVersionConstraint{}
	for _, requiredMod := range newRequire.Mods {
		if modInOld := oldRequire.GetModDependency(requiredMod.Name); modInOld == nil {
			modsToAdd = append(modsToAdd, requiredMod)
		}
	}

	if len(modsToAdd) == 0 {
		// an empty changeset
		return ChangeSet{}
	}

	// create the HCL serialization for the mod blocks which needs to be placed
	// in the require block
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	for _, modToAdd := range modsToAdd {
		rootBody.AppendBlock(i.createNewModRequireBlock(modToAdd))
	}

	return ChangeSet{
		&Change{
			Operation:   Insert,
			OffsetStart: oldRequire.BodyRange.End.Byte - 1,
			Content:     f.Bytes(),
		},
	}
}

// calculates the changes required in mod.sp to reflect updates
func (i *ModInstaller) calcChangesForUpdate(oldRequire *modconfig.Require, newRequire *modconfig.Require) ChangeSet {
	changes := ChangeSet{}
	for _, requiredMod := range oldRequire.Mods {
		modInUpdated := newRequire.GetModDependency(requiredMod.Name)
		if modInUpdated == nil {
			continue
		}
		if modInUpdated.VersionString != requiredMod.VersionString {
			changes = append(changes, &Change{
				Operation:   Replace,
				OffsetStart: requiredMod.VersionRange.Start.Byte,
				OffsetEnd:   requiredMod.VersionRange.End.Byte,
				Content:     []byte(fmt.Sprintf("version = \"%s\"", modInUpdated.VersionString)),
			})
		}
	}
	return changes
}
