dashboard "with_syntax" {

  with "data" {
    query = query.data
  }
#
#  table {
#    query = query.data
#  }
#
#  table {
#    args = [ with.data.rows[0].person, with.data.rows[0].server ]
#    sql = <<EOQ
#      select
#        $1 as person,
#        $2 as server
#    EOQ
#  }

  table {
    args = [ with.data.rows[*].person, with.data.rows[*].server ]
    sql = <<EOQ
      select
        $1 as person,
        $2 as server
    EOQ
  }

}

query "data" {
  sql = <<EOQ
    with data(person, server) as (
      values
        ('jon', 'mastodon.social'),
        ('chris', 'infosec.exchange')
    )
    select
      *
    from
      data
  EOQ
}