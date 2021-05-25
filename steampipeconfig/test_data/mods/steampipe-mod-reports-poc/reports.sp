report simple {
  title = "Simple"

  panel header {
    source = "steampipe.panel.markdown"
    text = "# Hello World"
  }
}

report two_panel {
  title = "Two Panel"

  panel hello {
    source = "steampipe.panel.markdown"
    text = "# Hello World"
    width = 6
  }

  panel goodbye {
    source = "steampipe.panel.markdown"
    text = "# Goodbye Universe"
    width = 6
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
