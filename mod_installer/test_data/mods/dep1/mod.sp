mod "dep1"{
  requires {
    plugin "azure"  {
      version = "v1.0"
    }
    mod "github.com/kaidaguerre/steampipe-mod-m2"  {
      version = "v1.0"
    }
  }
}
