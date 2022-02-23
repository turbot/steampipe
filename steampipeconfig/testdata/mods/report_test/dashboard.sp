dashboard "d" {
    container {
        chart {
            title = "dashboard_d_anonymous_container_0_anonymous_chart_0"
        }
        chart {
            title = "dashboard_d_anonymous_container_0_anonymous_chart_1"
            decription = chart.c.title

        }
    }
    container {
        chart {
            title = "dashboard_d_anonymous_container_1_anonymous_chart_0"
            decription = chart.c.title
        }
    }

}

chart "c"{
    decription = "foo"
}