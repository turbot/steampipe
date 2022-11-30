dashboard "failing_node" {
  title         = "failing_node"
  graph {
    title     = "Relationships"
    type      = "graph"
    //direction = "left_right" 

    nodes = [
      node.node_1,
      node.node_2,
      node.node_3,
      node.node_4,
      node.node_5,
    ]
    edges = [
      edge.edge_1,
      edge.edge_2,
      edge.edge_3,
      edge.edge_4,
    ]
  }
}

node "node_1" {
  sql = <<-EOQ
    select
      'node_1' as id,
      'node_1' as title
  EOQ
}
node "node_2" {
  sql = <<-EOQ
    select

      'node_2' as id,
      'node_2' as title
  EOQ
}
node "node_3" {
  sql = <<-EOQ
    select
      'node_3' as id,
      'node_3' as title
  EOQ
}
node "node_4" {
  sql = <<-EOQ
    select
      'node_4' as id,
      'node_4' as title
  EOQ
}
node "node_5" {
  sql = <<-EOQ
   select
      'node_5' as id,
      'node_5' as title
  EOQ

}

edge "edge_1" {
  sql = <<-EOQ
    select
      'node_1' as from_id,
      'node_2' as to_id
  EOQ
}
edge "edge_2" {
  sql = <<-EOQ
    select
      'node_1' as from_id,
      'node_3' as to_id
  EOQ
}
edge "edge_3" {
  sql = <<-EOQ
    select
ccccc
      'node_1' as from_id,
      'node_4' as to_id
  EOQ
}
edge "edge_4" {
  sql = <<-EOQ
    select
      'node_1' as from_id,
      'node_5' as to_id
  EOQ
}
