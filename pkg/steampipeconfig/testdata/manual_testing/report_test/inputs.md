# Top level input and dashboard scoped input

```

query "q1"{
    sql = "select 1"
    param "p1"{
    }
}

dashboard "r1"{
    input "i1"{ }

    chart {
        query = query.q1
        args = {
            "p1" = self.input.i1.value
        }
    }
}

dashboard "r2"{
    input "i1"{ }

    chart {
        query = query.q1
        args = {
            "p1" = self.input.i1.value
        }
    }
}
```

# Mod inputs

- local.dashboard.r1.input.i1
- local.dashboard.r2.input.i1

# Dashboard r1 inputs
- input.i1

- # Dashboard r2 inputs
- input.i1