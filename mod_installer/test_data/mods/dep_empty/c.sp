
benchmark "my_mod_public_resources" {
  title       = "Public Resources"
  description = "Resources that are public."
  children = [
    aws_compliance.benchmark.cis_v140_1,
  ]
}