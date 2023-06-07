# bad_mod_with_plugin_require_not_met

### Description

This mod is used to test that while running steampipe from the mod folder, the requirements mentioned in mod.sp `require` section are always respected.

### Usage

This mod is used in the tests in `mod_require.bats` to simulate a scenario where mod installation would fail because of a plugin version requirement not being satisfied.

Trying to install the mod would result in an error:
`Error: could not find plugin which satisfies requirement 'gcp@99.21.0' - required by 'bad_mod_with_require_not_met'`.

Running steampipe from this mod folder would throw a warning:
`Warning: could not find plugin which satisfies requirement 'gcp@99.21.0' - required by 'bad_mod_with_require_not_met'`