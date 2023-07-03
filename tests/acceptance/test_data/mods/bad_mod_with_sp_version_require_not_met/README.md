# bad_mod_with_sp_version_require_not_met

### Description

This mod is used to test that while running steampipe from the mod folder, the requirements mentioned in mod.sp `require` section are always respected.

### Usage

This mod is used in the tests in `mod_require.bats` to simulate a scenario where mod installation would fail because of steampipe CLI version requirement not being satisfied.

Trying to install the mod would result in an error:
`Error: steampipe version x.x.x does not satisfy mod.bad_mod_with_sp_version_require_not_met which requires version 10.99.99`.

Running steampipe from this mod folder would throw a warning:
`Warning: steampipe version x.x.x does not satisfy mod.bad_mod_with_sp_version_require_not_met which requires version 10.99.99`