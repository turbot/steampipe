# dashboard_parsing_nested_node_edge_providers_fail

### Description

This is a mod containing a dashboard with flow, hierarchy and graph(node and edge providers). This is used for testing hcl parsing - nested Node and Edge providers always require a query/sql block or a node/edge block. Running `steampipe dashboard` from this mod would result in parsing failures.

### Usage

This mod is used in the tests in `dashboard_parsing_validation.bats` to test the dashboard hcl parsing functionality. This mod has flow, hierarchy and graph which require either a query/sql block or a node/edge block. Run `steampipe dashboard` to get the parsing validation error.