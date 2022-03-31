dashboard "control" {
    title = "WRAPPED"

  benchmark {
    base = aws_compliance.benchmark.cis_v140
  }

}

benchmark "tl" {
    title = "TOP LEVEL"
    base = aws_compliance.benchmark.cis_v140
}