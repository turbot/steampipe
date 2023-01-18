dashboard "Benchmarks" {

  benchmark {
    base = benchmark.my_cis_v140
  }

}

benchmark "my_cis_v140" {
  title         = "ACME CIS v1.4.0"
  description   = "Only sections 1, 2, and 3."
  children = [
    aws_compliance.benchmark.cis_v140_1,
    aws_compliance.benchmark.cis_v140_2,
    aws_compliance.benchmark.cis_v140_3
  ]
}