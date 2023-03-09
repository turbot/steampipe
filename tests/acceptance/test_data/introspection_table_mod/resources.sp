variable "sample_var_1"{
	type = string
	default = "steampipe_var"
}


query "sample_query_1"{
	title ="Sample query 1"
	description = "query 1 - 3 params all with defaults"
	sql = "select 'ok' as status, 'steampipe' as resource, concat($1::text, $2::text, $3::text) as reason"
	param "p1"{
			description = "p1"
			default = var.sample_var_1
	}
	param "p2"{
			description = "p2"
			default = "because_def "
	}
	param "p3"{
			description = "p3"
			default = "string"
	}
}

control "sample_control_1" {
  title = "Sample control 1"
  description = "Sample control to test introspection functionality"
  query = query.sample_query_1
  severity = "high"
  tags = {
    "foo": "bar"
  }
}

benchmark "sample_benchmark_1" {
	title = "Sample benchmark 1"
	description = "Sample benchmark to test introspection functionality"
	children = [
		control.sample_control_1
	]
}

dashboard "sample_dashboard_1" {
  title = "Sample dashboard 1"
  description = "Sample dashboard to test introspection functionality"

  container "sample_conatiner_1" {
		card "sample_card_1" {
			title = "Sample card 1"
		}

		image "sample_image_1" {
			title = "Sample image 1"
			width = 3
  		src = "https://steampipe.io/images/logo.png"
  		alt = "steampipe"
		}

		text "sample_text_1" {
			title = "Sample text 1"
		}

    chart "sample_chart_1" {
      sql = "select 1 as chart"
      width = 5
      title = "Sample chart 1"
    }

    flow "sample_flow_1" {
      title = "Sample flow 1"
      width = 3

      node "sample_node_1" {
        sql = <<-EOQ
          select 1 as node
        EOQ
      }
      edge "sample_edge_1" {
        sql = <<-EOQ
          select 1 as edge
        EOQ
      }
    }

    graph "sample_graph_1" {
      title = "Sample graph 1"
      width = 5

      node "sample_node_2" {
        sql = <<-EOQ
          select 1 as node
        EOQ
      }
      edge "sample_edge_2" {
        sql = <<-EOQ
          select 1 as edge
        EOQ
      }
    }

    hierarchy "sample_hierarchy_1" {
      title = "Sample hierarchy 1"
      width = 5

      node "sample_node_3" {
        sql = <<-EOQ
          select 1 as node
        EOQ
      }
      edge "sample_edge_3" {
        sql = <<-EOQ
          select 1 as edge
        EOQ
      }
    }

    table "sample_table_1" {
      sql = "select 1 as table"
      width = 4
      title = "Sample table 1"
    }

    input "sample_input_1" {
      sql = "select 1 as input"
      width = 2
      title = "Sample input 1"
    }
  }
}
