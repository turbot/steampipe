variable local_with_default {
  default = "select 'local with default' as a"
}

variable local_set_in_file {

}


variable unset {
  type = string
  description = "ooh something or other"
}

variable dupe_name_var {
  type = string
}

query local_with_default{
  sql = var.local_with_default
}

query dupe_name_var{
  sql = var.dupe_name_var
}

query base_dupe_name_var{
  sql = m1.var.dupe_name_var
}

query dep_mod_var1{
  sql = m1.var.dep_mod_var1
}

query dep_mod_var2{
  sql = m1.var.dep_mod_var2
}

query local_set_in_file{
  sql = var.local_set_in_file
}
