# Workspace Profile (work in progress)

## Adding properties to Workspace Profile

### Adding simple properties to `Workspace Profile`

* Add properties to the `WorkspaceProfile` struct in `pkg/steampipeconfig/modconfig/workspace_profile.go`.
* Add `hcl` and `cty` tags to the properties. (eample: `hcl:"search_path" cty:"search_path"`).
* Add to `(p *WorkspaceProfile) setBaseProperties()`. This enables `base` profile inheritance. **Remember to check for `nil`**.
* Add to `(p *WorkspaceProfile) ConfigMap(commandName string)`.

### Adding an `options` property. [Example Commit](https://github.com/turbot/steampipe/pull/3228/commits/642f6fd20cf98aed2e2ab393a9d86345b53872a1)

#### Define `struct` with the following interface

```
type Query struct {}

// ConfigMap :: this is merged with viper
// Only add keys which are not nil
func (t *Query) ConfigMap() map[string]interface{} {}

// Merge :: merge other options over the top of this options object
// i.e. if a property is set in otherOptions, it takes precedence
func (t *Query) Merge(otherOptions Options) {
  // make sure this is the type we want
  if _, ok := otherOptions.(*Query); !ok {
		return
	}
}

// String serialize for printing
func (t *Query) String() string {}
```

#### Add `struct` tags

For properties in the struct which need to be extracted from the HCL, add the following tag

```
hcl:"output"
```

where `output` is the property in the HCL.

#### Add to `pkg/steampipeconfig/parse/decode_options.go`
#### Add to `pkg/steampipeconfig/options/options.go`
#### Add to `pkg/steampipeconfig/modconfig/workspace_profile.go` in `WorkspaceProfile` struct
##### Update `(p *WorkspaceProfile) SetOptions` in `pkg/steampipeconfig/modconfig/workspace_profile.go`
##### Update `(p *WorkspaceProfile) ConfigMap(commandName string)` in `pkg/steampipeconfig/modconfig/workspace_profile.go`