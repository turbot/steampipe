mod "m1"{

  requires{
    steampipe = "v0.8.0"
    mod "github.com/turbot/aws-core"  {
      version = "v1.0"
    }
  }
}