
benchmark "c1"{

}
control "c2"{
    parent = benchmark.c1
}
benchmark "c3"{
    parent= benchmark.c1
}