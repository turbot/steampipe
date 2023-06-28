# bad_mod_with_dep_mod_version_require_not_met

### Description

This mod is used to test that while running steampipe from the mod folder, the requirements mentioned in mod.sp `require` section are always respected.

### Usage

This mod is used in the tests in `mod_require.bats` to simulate a scenario where mod installation would fail because of a dependant mod version requirement not being satisfied.

Trying to install the mod would result in an error:
`Error: 1 dependency failed to install - no version of github.com/turbot/steampipe-mod-aws-compliance found satisfying version constraint: 99.21.0`.
