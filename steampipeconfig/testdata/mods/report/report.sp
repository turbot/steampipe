chart "c1"{
  type = "bar"
  title = "Unencrypted and Non-Versioned Buckets by Region"
  legend {
    display  = "auto"
    position = "top"
  }
  series other {
    title = "Configured Buckets"
    color = "green"
  }
  series unencrypted {
    title = "Unencrypted Buckets"
    color = "red"
  }
  series nonversioned {
    title = "Non-Versioned Buckets"
    color = "orange"
  }
  axes {
    x {
      title = "Regions"
      labels {
        display = "auto"
      }
    }
    y {
      title  = "Totals"
      labels {
        display = "show"
      }
      min    = 0
      max    = 100
      steps  = 10
    }
  }
}