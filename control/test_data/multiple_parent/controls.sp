

benchmark "b1" {
  title         = "1 Identity and Access Management"
  children = [
    control.cis_v130_1_1,
    control.cis_v130_1_2,
  ]
}

benchmark "b2" {
  title         = "1 Identity and Access Management"
  children = [
    control.cis_v130_1_1,
    control.cis_v130_1_2,
  ]
}

control "cis_v130_1_1" {
  title         = "1.1 Maintain current contact details"
  sql           = query.alarm.sql
}

control "cis_v130_1_2" {
  title         = "1.2 Ensure security contact information is registered"
  sql           = query.alarm.sql
}
