dashboard "my_dash" {
  with "dash_level" {...}
  with "duplicate_name_at_each_level" {...}
  with "duplicate_name_at_2_levels" {...}

  graph "ec2_instance_detail" {
    with "graph_level" {...}
    with "duplicate_name_at_each_level" {...}
    with "duplicate_name_at_2_levels" {...}

    node {
      with "node_level" {}
      with "duplicate_name_at_each_level" {...}

      args = {
        a = with.dash_level.rows[*].foo   # dashboard level
        b = with.graph_level.rows[*].foo  # graph level
        c = with.node_level.rows[*].foo   # node level
        d = with.duplicate_name_at_each_level.rows[*].foo  # node level wins
        e = with.duplicate_name_at_2_levels.rows[*].foo  # graph level wins
      }
    }

    node {
      base = node.top_level_node
      args = {
        a = with.dash_level.rows[*].foo   # dashboard level
        b = with.graph_level.rows[*].foo  # graph level
        d = with.duplicate_name_at_each_level.rows[*].foo  # graph level wins
        # - the node level with.node_level is not visible here because its defined
        # in the base's namespace, not here
        # Likewise, cannot reference with.node_level in hcl here
      }

    }
  }


  node "top_level_node" {
    with "node_level" {}
    with "duplicate_name_at_each_level" {...}

    args = {
      # cannot refer to dashboard level or graph level `with` because
      # this is a top level node
      c = with.node_level.rows[*].foo   # node level
      d = with.duplicate_name_at_each_level.rows[*].foo  # node level
    }
  }