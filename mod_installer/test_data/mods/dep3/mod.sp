mod "dep3"{
  requires {
    mod "github.com/kaidaguerre/steampipe-mod-m1"  {
      version = "v1.*"
    }
    mod "github.com/kaidaguerre/steampipe-mod-m2"  {
      version = "v3.1"
    }
  }
}
