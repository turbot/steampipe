dashboard "nodes_edges" {
  text {
    value = "a simple report"
  }

 graph {
   type  = "graph"

   args = {
     arn = "FOO"
   }
   nodes = [node.n1]
   node = node.n1
 }
}

node "n1"{
    title = "n1"
    sql = "select 1"
}
