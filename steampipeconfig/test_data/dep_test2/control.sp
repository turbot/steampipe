
control_group "c1"{

}
control "c2"{
    parent = control_group.c1
}
control_group "c3"{
    parent= control_group.c1
}