

chart aws_bucket_info {
  type = "column"
  grouping = "compare"
  legend {
    position = "bottom"
  }
  axes {
    x {
      title {
        display = "always"
        value = "Foo"
      }
    }
    y {
      title {
        display = "always"
        value = "Foo"
      }
    }
  }
}

report override_base_values {
  title = "override_base_values"

  chart {
    base = chart.aws_bucket_info
    axes {
        x {
          title {
            value = "OVERRIDE"
          }
        }
        y {
          title {
            display = "OVERRIDE"
          }
        }
    }
  }
}