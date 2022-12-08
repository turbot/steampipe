dashboard "testing_card_blocks" {
  title = "Testing card blocks"

  container {
    card "card1" {
      sql = <<-EOQ
        select 1 as card1_value
      EOQ
      width = 2
    }

    card "card2" {
      type  = "info"
      width = 2
      sql = <<-EOQ
        select 2 as card2_value
      EOQ
    }
  }
}