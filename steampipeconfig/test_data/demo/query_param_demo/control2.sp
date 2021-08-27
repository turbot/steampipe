
variable "prohibited_instance_types" {
    type    = list(string)
    default = ["%4xl","%8xl","%12xl","%16xl","%24xl","%32xl","%.metal"]
}

control "array_param" {
    title       = "EC2 Instances xlarge and bigger"
    params  = [ var.prohibited_instance_types ]
    query         = query.q2
}
