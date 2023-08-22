# Mods Index

| Name | Description |
|------|-------------|
| [bad_mod_with_dep_mod_version_require_not_met](bad_mod_with_dep_mod_version_require_not_met/README.md) |  This mod is used to test that while running steampipe from the mod folder, the requirements mentioned in mod.sp `require` section are always respected. |
| [bad_mod_with_plugin_require_not_met](bad_mod_with_plugin_require_not_met/README.md) |  This mod is used to test that while running steampipe from the mod folder, the requirements mentioned in mod.sp `require` section are always respected. |
| [bad_mod_with_sp_version_require_not_met](bad_mod_with_sp_version_require_not_met/README.md) |  This mod is used to test that while running steampipe from the mod folder, the requirements mentioned in mod.sp `require` section are always respected. |
| [check_all_mod](check_all_mod/README.md) |  This mod is used to test the `check all` functionality. |
| [config_parsing_test_mod](config_parsing_test_mod/README.md) |  This is a simple mod used for testing the steampipe connection config parsing. This mod will ONLY work in acceptance tests. |
| [control_rendering_test_mod](control_rendering_test_mod/README.md) |  This is a simple mod used for testing the steampipe check output and exports rendering. This mod is used in acceptance tests. |
| [dashboard_cards](dashboard_cards/README.md) |  This is a simple mod containing a dashboard with cards. This is used for testing dashboard card mod resource. |
| [dashboard_graphs](dashboard_graphs/README.md) |  This is a simple mod containing a dashboard with graphs, nodes and edges. This is used for testing dashboard graph, node and edge mod resources. |
| [dashboard_inputs](dashboard_inputs/README.md) |  This is a simple mod containing a dashboard with inputs. This is used for testing dashboard text inputs. |
| [dashboard_parsing_nested_node_edge_providers_fail](dashboard_parsing_nested_node_edge_providers_fail/README.md) |  This is a mod containing a dashboard with flow, hierarchy and graph(node and edge providers). This is used for testing hcl parsing - nested Node and Edge providers always require a query/sql block or a node/edge block. Running `steampipe dashboard` from this mod would result in parsing failures. |
| [dependent_mod_with_legacy_lock](dependent_mod_with_legacy_lock/README.md) |  This is a test mod with legacy lock. Run `steampipe mod install` to install the dependent mods in this folder, before running the tests. |
| [dependent_mod_with_variables](dependent_mod_with_variables/README.md) |  This is a test mod which depends on another mod. Run `steampipe mod install` to install the dependent mods in this folder, before running the tests. |
