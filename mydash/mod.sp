mod "test_compliance" {
  # hub metadata
  title       = "Test Compliance"
  description = "Test Compliance"
}

variable "string_list" {
  type        = list(string)
  default     = []
}

variable "number_list" {
  type        = list(number)
  default     = []
}

variable "bool_list" {
  type        = list(bool)
  default     = []
}

query "string_list" {
  sql         = <<-EOQ
    select ($1::text[]) as string_list
  EOQ

  param "string_list_param" {
    default = var.string_list
  }
}

query "number_list" {
  sql         = <<-EOQ
    select ($1::text[]) as number_list
  EOQ

  param "number_list_param" {
    default = var.number_list
  }
}

query "bool_list" {
  sql         = <<-EOQ
    select ($1::text[]) as bool_list
  EOQ

  param "bool_list_param" {
    default = var.bool_list
  }
}
