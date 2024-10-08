package constants

// DefaultWorkspaceContent is the content of the sample workspaces config file(workspaces.spc.sample),
// that is created if it does not exist
const DefaultWorkspaceContent = `
#
# For detailed descriptions, see the reference documentation
# at https://steampipe.io/docs/reference/config-files/workspace
#

# workspace "all_options" {
#   pipes_host         = "pipes.turbot.com"
#   pipes_token        = "spt_999faketoken99999999_111faketoken1111111111111"
#   install_dir        = "~/steampipe2"
#   mod_location       = "~/src/steampipe-mod-aws-insights"  
#   query_timeout      = 300
#   snapshot_location  = "acme/dev"
#   workspace_database = "local" 
#   search_path        = "aws,aws_1,aws_2,gcp,gcp_1,gcp_2,slack,github"
#   search_path_prefix = "aws_all"
#   watch              = true
#   max_parallel       = 5
#   introspection      = false
#   input              = true
#   progress           = true
#   theme              = "dark"  # light, dark, plain 
#   cache              = true
#   cache_ttl          = 300
# 
# 
#   options "query" {
#     autocomplete = true
#     header       = true    # true, false
#     multi        = false   # true, false
#     output       = "table" # json, csv, table, line
#     separator    = ","     # any single char
#     timing       = on   # off, on, verbose
#   }
# }
`
