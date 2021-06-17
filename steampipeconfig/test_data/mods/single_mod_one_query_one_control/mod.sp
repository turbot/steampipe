mod "m1"{
  title = "M1"
  description = "THIS IS M1"
  requires{
    plugin "aws"{
      version = "0.24.0"
    }
    plugin "gcp"{
      version = "0.12.0"
    }
    plugin "turbot/chaos"{
      version = "0.11.0"
    }
  }
}