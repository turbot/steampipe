package mod_installer

import (
	//"fmt"
	//
	//"github.com/xlab/treeprint"
	_ "github.com/xlab/treeprint"
)

func (i *ModInstaller) GetModList() interface{} {
	//// install first if necessary (or error)
	//if len(i.workspaceLock) == 0 {
	//	if err := i.InstallWorkspaceDependencies(); err != nil {
	//		return err
	//	}
	//}
	//
	//// to add a custom root name use `treeprint.NewWithRoot()` instead
	//tree := treeprint.New()
	//
	//modName := i.workspaceMod.Name()
	//tree = i.addModToTree(tree, modName)
	//
	//tree.AddNode("Makefile")
	//tree.AddNode("aws.sh")
	//tree.AddMetaBranch(" 204", "bin").
	//	AddNode("dbmaker").AddNode("someserver").AddNode("testtool")
	//tree.AddMetaBranch(" 374", "deploy").
	//	AddNode("Makefile").AddNode("bootstrap.sh")
	//tree.AddMetaNode("122K", "testtool.a")
	//
	//fmt.Println(tree.String())
	return nil
}

//func (i *ModInstaller) addModToTree(tree treeprint.Tree, modName string) treeprint.Tree {
//	t := tree.AddBranch(modName)
//	for name, version := range i.workspaceLock[modName] {
//		t.AddNode(dep)
//	}
//}
