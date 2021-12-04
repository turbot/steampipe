mod "dep2"{
  requires {
    mod "github.com/kaidaguerre/steampipe-mod-m1"  {
      version = "v1.0"
    }
    mod "github.com/kaidaguerre/steampipe-mod-m2"  {
      version = "v3.0"
    }
  }
}
