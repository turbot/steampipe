query "query_params_with_all_defaults"{
  description = "query 1 - 3 params all with defaults"
  sql = "select 'ok' as status, 'steampipe' as resource, concat($1::text, ' ', $2::text, ' ', $3::text) as reason"
  param "p1"{
    description = "First parameter"
    default = "default_parameter_1"
  }
  param "p2"{
    description = "Second parameter"
    default = "default_parameter_2"
  }
  param "p3"{
    description = "Third parameter"
    default = "default_parameter_3"
  }
}

query "query_params_with_no_defaults"{
  description = "query 1 - 3 params with no defaults"
  sql = "select 'ok' as status, 'steampipe' as resource, concat($1::text, ' ', $2::text, ' ', $3::text) as reason"
  param "p1"{
    description = "First parameter"
  }
  param "p2"{
    description = "Second parameter"
  }
  param "p3"{
    description = "Third parameter"
  }
}

query "query_array_params_with_default"{
  description = "query an array parameter with default"
  sql = "select 'ok' as status, 'steampipe' as resource, $1::jsonb->1 as reason"
  param "p1"{
    description = "Array parameter"
    default = ["default_p1_element_01", "default_p1_element_02", "default_p1_element_03"]
  }
}

query "query_map_params_with_default"{
  description = "query a map parameter with default"
  sql = "select 'ok' as status, 'steampipe' as resource, $1::json->'default_property_01' as reason"
  param "p1"{
    description = "Map parameter"
    default = {"default_property_01": "default_property_value_01", "default_property_02": "default_property_value_02"}
  }
}

query "query_map_params_with_no_default"{
  description = "query a map parameter with no default"
  sql = "select 'ok' as status, 'steampipe' as resource, $1::json->'default_property_01' as reason"
  param "p1"{
    description = "Map parameter"
  }
}