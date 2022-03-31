dashboard "control" {
    title = "Steampipe Community [${formatdate("D-MMM-YYYY", timestamp())}] (with filter)"

  control {
    base = aws_compliance.control.cis_v140_1_5
  }

  benchmark {
    base = aws_compliance.benchmark.cis_v140
  }

}

benchmark "tl" {
    title = "MY BENCHMARK"
    base = aws_compliance.benchmark.cis_v140
}