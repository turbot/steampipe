dashboard "community_filter" {
    title = "Steampipe Community [${formatdate("D-MMM-YYYY", timestamp())}] (with filter)"

    input "search" {
        title = "community_filter"
        width       = 8
        type        = "text"
        label       = "Search"
        placeholder = "Enter search phrase..."
    }
}
dashboard "community_filter2" {
    title = "Steampipe Community 2"

    input "search2" {
        title = "community_filter2"
        width       = 8
        type        = "text"
        label       = "Search"
        placeholder = "Enter search phrase..."
    }
}
dashboard "community_filter3" {
    title = "Steampipe Community 3)"

    input "search" {
        title = "community_filter3"
        width       = 8
        type        = "text"
        label       = "Search"
        placeholder = "Enter search phrase..."
    }
}