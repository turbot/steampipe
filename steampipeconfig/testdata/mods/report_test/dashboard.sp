
chart "top_level1"{}
chart "top_level2"{}

dashboard "anonymous_naming" {

    chart{}

    container {
        chart {}
        chart {}
        table{}
    }

    container {
        chart {}
        chart {}
        table{}
        table{}
    }
}
