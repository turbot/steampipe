report simple {
  title = "Simple"

  panel header {
    source = "steampipe.panel.markdown_"
    text = "# Hello World"
  }
}

report two_panel {
  title = "Two Panel"

  panel hello {
    source = "steampipe.panel.markdown"
    text = "# Hello World"
    width = 6
    height = 1
  }

  panel goodbye {
    source = "steampipe.panel.markdown"
    text = "# Goodbye Universe"
    width = 6
    height = 2
  }
}

report barchart {
  title = "Bar Chart"
  
  panel header {
    source = "steampipe.panel.markdown"
    text = "# Simple Bar Chart"
  }

  panel barchart {
    title = "AWS IAM Entities"
    source = "steampipe.panel.barchart"
    sql = query.aws_iam_entities.sql
  }
}
