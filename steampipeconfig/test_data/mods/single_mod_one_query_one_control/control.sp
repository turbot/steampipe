control "c1"{
    title ="C1"
    description = "THIS IS CONTROL 1"
    tags = {
    "Application" = "demo"
    "EnvironmentType" = "prod"
    "ProductType" = "steampipe"
}
    sql = "select 'pass' as result"
}

control "c2"{
    title ="C2"
    description = "THIS IS CONTROL 2"
    tags = {
    "Application" = "demo"
    "EnvironmentType" = "prod"
    "ProductType" = "steampipe"
}
    sql = "select 'pass' as result"
}

control "c3"{
    title ="C3"
    description = "THIS IS CONTROL 3"
    tags = {
    "Application" = "demo"
    "EnvironmentType" = "prod"
    "ProductType" = "steampipe"
}
    sql = "select 'fail' as result"
}

control "c4"{
    title ="C4"
    description = "THIS IS CONTROL 4"
    tags = {
    "Application" = "demo"
    "EnvironmentType" = "prod"
    "ProductType" = "steampipe"
}
    sql = "select 'pass' as result"
}
