
ModParseContext has LoadedDependencyMods modconfig.ModMap

currently keyed by mod name - change to key by full name of locked version

GetLockedModVersionConstraint()
FullName()

Usage

1) loadModDependencies
```go 
func loadModDependencies(mod *modconfig.Mod, parseCtx *parse.ModParseContext) error {
    ...
    for _, requiredModVersion := range mod.Require.Mods {
        // if we have a locked version, update the required version to reflect this
        lockedVersion, err := parseCtx.WorkspaceLock.GetLockedModVersionConstraint(requiredModVersion, mod)
        if err != nil {
            errors = append(errors, err)
            continue
        }
        if lockedVersion != nil {
            requiredModVersion = lockedVersion
        }

        // have we already loaded a mod which satisfied this
        if loadedMod, ok := parseCtx.LoadedDependencyMods[requiredModVersion.Name]; ok {

```