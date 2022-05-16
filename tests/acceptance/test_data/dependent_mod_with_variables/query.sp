variable local_with_default {
  default = "select 'local with default' as a"
}

variable local_set_in_file {

}

#variable dupe_name_var {
#  type = string
#}

query local_with_default{
  sql = var.local_with_default
}
#
#query dupe_name_var{
#  sql = var.dupe_name_var
#}

query local_set_in_file{
  sql = var.local_set_in_file
}
