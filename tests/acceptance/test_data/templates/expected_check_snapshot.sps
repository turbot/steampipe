{
    "end_time": "2022-12-15T20:12:43.270226+05:30",
    "inputs": {},
    "layout": {
        "name": "control_rendering_test_mod.control.sample_control_mixed_results_1",
        "panel_type": "control"
    },
    "panels": {
        "control_rendering_test_mod.control.sample_control_mixed_results_1": {
            "data": {
                "columns": [
                    {
                        "data_type": "TEXT",
                        "name": "reason"
                    },
                    {
                        "data_type": "TEXT",
                        "name": "resource"
                    },
                    {
                        "data_type": "TEXT",
                        "name": "status"
                    },
                    {
                        "data_type": "INT4",
                        "name": "id"
                    }
                ],
                "rows": [
                    {
                        "id": "16",
                        "reason": "Resource has some error",
                        "resource": "steampipe",
                        "status": "error"
                    },
                    {
                        "id": "17",
                        "reason": "Resource has some error",
                        "resource": "steampipe",
                        "status": "error"
                    },
                    {
                        "id": "11",
                        "reason": "Resource does not satisfy condition",
                        "resource": "steampipe",
                        "status": "alarm"
                    },
                    {
                        "id": "12",
                        "reason": "Resource does not satisfy condition",
                        "resource": "steampipe",
                        "status": "alarm"
                    },
                    {
                        "id": "13",
                        "reason": "Resource does not satisfy condition",
                        "resource": "steampipe",
                        "status": "alarm"
                    },
                    {
                        "id": "14",
                        "reason": "Resource does not satisfy condition",
                        "resource": "steampipe",
                        "status": "alarm"
                    },
                    {
                        "id": "15",
                        "reason": "Resource does not satisfy condition",
                        "resource": "steampipe",
                        "status": "alarm"
                    },
                    {
                        "id": "19",
                        "reason": "Information",
                        "resource": "steampipe",
                        "status": "info"
                    },
                    {
                        "id": "20",
                        "reason": "Information",
                        "resource": "steampipe",
                        "status": "info"
                    },
                    {
                        "id": "21",
                        "reason": "Information",
                        "resource": "steampipe",
                        "status": "info"
                    },
                    {
                        "id": "1",
                        "reason": "Resource satisfies condition",
                        "resource": "steampipe",
                        "status": "ok"
                    },
                    {
                        "id": "2",
                        "reason": "Resource satisfies condition",
                        "resource": "steampipe",
                        "status": "ok"
                    },
                    {
                        "id": "3",
                        "reason": "Resource satisfies condition",
                        "resource": "steampipe",
                        "status": "ok"
                    },
                    {
                        "id": "4",
                        "reason": "Resource satisfies condition",
                        "resource": "steampipe",
                        "status": "ok"
                    },
                    {
                        "id": "5",
                        "reason": "Resource satisfies condition",
                        "resource": "steampipe",
                        "status": "ok"
                    },
                    {
                        "id": "6",
                        "reason": "Resource satisfies condition",
                        "resource": "steampipe",
                        "status": "ok"
                    },
                    {
                        "id": "7",
                        "reason": "Resource satisfies condition",
                        "resource": "steampipe",
                        "status": "ok"
                    },
                    {
                        "id": "8",
                        "reason": "Resource satisfies condition",
                        "resource": "steampipe",
                        "status": "ok"
                    },
                    {
                        "id": "9",
                        "reason": "Resource satisfies condition",
                        "resource": "steampipe",
                        "status": "ok"
                    },
                    {
                        "id": "10",
                        "reason": "Resource satisfies condition",
                        "resource": "steampipe",
                        "status": "ok"
                    },
                    {
                        "id": "18",
                        "reason": "Resource is skipped",
                        "resource": "steampipe",
                        "status": "skip"
                    }
                ]
            },
            "description": "Sample control that returns 10 OK, 5 ALARM, 2 ERROR, 1 SKIP and 3 INFO",
            "name": "control_rendering_test_mod.control.sample_control_mixed_results_1",
            "panel_type": "control",
            "properties": {
                "name": "sample_control_mixed_results_1",
                "severity": "high"
            },
            "status": "complete",
            "summary": {
                "alarm": 5,
                "error": 2,
                "info": 3,
                "ok": 10,
                "skip": 1
            },
            "title": "Sample control with all possible statuses(severity=high)"
        }
    },
    "schema_version": "20220929",
    "start_time": "2022-12-15T20:12:43.263569+05:30",
    "variables": {}
}