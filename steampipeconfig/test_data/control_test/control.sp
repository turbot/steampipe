control "cis_v130_1_1" {
    title         = "1.1 - Maintain current contact details (Manual)"
    description   = "Ensure contact email and telephone details for AWS accounts are current and map to more than one individual in your organization."
    sql           = query.manual_control.sql
}
