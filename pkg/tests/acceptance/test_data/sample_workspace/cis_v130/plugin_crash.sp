benchmark "check_plugin_crash_benchmark" {
  title         = "Benchmark to test the plugin crash bug while running controls"
  children = [
    control.plugin_chaos_test_1,
    control.plugin_crash_test,
    control.plugin_chaos_test_2
  ]
}

control "plugin_chaos_test_1" {
  title       = "Control to query a chaos table"
  description = "Control to query a chaos table to test all flavours of integer and float data types"
  sql         = query.check_plugincrash_normalquery1.sql
  severity    = "high"
}

control "plugin_crash_test" {
  title       = "Control to simulate a plugin crash"
  description = "Control to query a chaos table that prints 50 rows and do an os.Exit(-1) to simulate a plugin crash"
  sql         = "select * from chaos_plugin_crash"
  severity    = "high"
}

control "plugin_chaos_test_2" {
  title       = "Control to query a chaos table"
  description = "Control to query a chaos table test the Get call with all the possible scenarios like errors, panics and delays"
  sql         = query.check_plugincrash_normalquery2.sql
  severity    = "high"
}