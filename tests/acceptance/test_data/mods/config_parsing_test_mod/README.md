# config_parsing_test_mod

### Description

This is a simple mod used for testing the steampipe connection config parsing. This mod will ONLY work in acceptance tests.

### Usage

This mod is used in the tests in `cache.bats` to test the steampipe connection config parsing functionality. The query in this mod uses the `chaos6` connection from `tests/acceptance/test_data/source_files/chaos_options.spc` which is used to verify caching. 