
variable "prohibited_instance_types" {
    type    = map
    default = {
        a = "foo"
    }
}

control "array_param" {
    title       = "EC2 Instances xlarge and bigger"
    args  = [ var.prohibited_instance_types ]
    query         = query.q2
}
