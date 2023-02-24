connection "con1"{
  plugin = "chaosdynamic"
  tables = [
    {
      name    = "t1"
      description = "test table 1"
      columns = [
        {
          name = "c1"
          type = "string"
        },
         {
          name = "c2"
          type = "string"
        }
      ]
    }
  ]
}

connection "con2"{
  plugin = "chaosdynamic"
  tables = [
    {
      name    = "t1"
      description = "test table 1"
      columns = [
        {
          name = "c1"
          type = "string"
        },
         {
          name = "c2"
          type = "ipaddr"
        }
      ]
    }
  ]
}

connection "dyn_agg"{
  plugin = "chaosdynamic"
  type = "aggregator"
  connections = ["con1", "con2"]
}