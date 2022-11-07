import { DashboardExecutionEventWithSchema, PanelDefinition } from "../types";
import { stripSnapshotDataForExport } from "./snapshot";

describe("snapshot utils", () => {
  describe("stripSnapshotDataForExport", () => {
    test("Schema 20220614", () => {
      const inputSnapshot: DashboardExecutionEventWithSchema = {
        schema_version: "20220614",
        execution_id: "0x140029247e0",
        dashboard_node: {
          name: "aws_insights.dashboard.aws_iam_user_dashboard",
        },
        layout: {
          name: "aws_insights.dashboard.aws_iam_user_dashboard",
          panel_type: "dashboard",
          children: [
            {
              name: "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0",
              panel_type: "container",
            },
          ],
        },
        panels: {
          "aws_insights.dashboard.aws_iam_user_dashboard": {
            name: "aws_insights.dashboard.aws_iam_user_dashboard",
            documentation: "# Some documentation",
            sql: "select something from somewhere",
            source_definition: 'some { hcl: "values" }',
            properties: {
              search_path: ["some_schema"],
              search_path_prefix: ["some_prefix"],
              sql: "select something from somewhere",
            },
          },
          "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0":
            {
              name: "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0",
              documentation: "# Some documentation",
              sql: "select something from somewhere",
              source_definition: 'some { hcl: "values" }',
              properties: {
                search_path: ["some_schema"],
                search_path_prefix: ["some_prefix"],
                sql: "select something from somewhere",
              },
            },
        },
        inputs: {
          "input.foo": "bar",
        },
        variables: {
          foo: "bar",
        },
        search_path: ["some_schema"],
        search_path_prefix: ["some_prefix"],
        start_time: "2022-10-27T14:43:57.79514+01:00",
        end_time: "2022-10-27T14:43:58.045925+01:00",
      };

      const migratedEvent = stripSnapshotDataForExport(inputSnapshot);

      const expectedPanels = {};

      for (const [name, panel] of Object.entries(inputSnapshot.panels)) {
        const { documentation, sql, source_definition, ...rest } =
          panel as PanelDefinition;
        expectedPanels[name] = { ...rest, properties: {} };
      }

      const expectedEvent = {
        schema_version: inputSnapshot.schema_version,
        dashboard_node: inputSnapshot.dashboard_node,
        execution_id: inputSnapshot.execution_id,
        layout: inputSnapshot.layout,
        panels: expectedPanels,
        inputs: inputSnapshot.inputs,
        variables: inputSnapshot.variables,
        start_time: inputSnapshot.start_time,
        end_time: inputSnapshot.end_time,
      };

      expect(migratedEvent).toEqual(expectedEvent);
    });

    test("Schema 20220929", () => {
      const inputSnapshot: DashboardExecutionEventWithSchema = {
        schema_version: "20220929",
        layout: {
          name: "aws_insights.dashboard.aws_iam_user_dashboard",
          panel_type: "dashboard",
          children: [
            {
              name: "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0",
              panel_type: "container",
            },
          ],
        },
        panels: {
          "aws_insights.dashboard.aws_iam_user_dashboard": {
            name: "aws_insights.dashboard.aws_iam_user_dashboard",
            documentation: "# Some documentation",
            sql: "select something from somewhere",
            source_definition: 'some { hcl: "values" }',
            properties: {
              search_path: ["some_schema"],
              search_path_prefix: ["some_prefix"],
              sql: "select something from somewhere",
            },
          },
          "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0":
            {
              name: "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0",
              documentation: "# Some documentation",
              sql: "select something from somewhere",
              source_definition: 'some { hcl: "values" }',
              properties: {
                search_path: ["some_schema"],
                search_path_prefix: ["some_prefix"],
                sql: "select something from somewhere",
              },
            },
        },
        inputs: {
          "input.foo": "bar",
        },
        variables: {
          foo: "bar",
        },
        search_path: ["some_schema"],
        search_path_prefix: ["some_prefix"],
        start_time: "2022-10-27T14:43:57.79514+01:00",
        end_time: "2022-10-27T14:43:58.045925+01:00",
      };

      const migratedEvent = stripSnapshotDataForExport(inputSnapshot);

      const expectedPanels = {};

      for (const [name, panel] of Object.entries(inputSnapshot.panels)) {
        const { documentation, sql, source_definition, ...rest } =
          panel as PanelDefinition;
        expectedPanels[name] = { ...rest, properties: {} };
      }

      const expectedEvent = {
        schema_version: inputSnapshot.schema_version,
        layout: inputSnapshot.layout,
        panels: expectedPanels,
        inputs: inputSnapshot.inputs,
        variables: inputSnapshot.variables,
        start_time: inputSnapshot.start_time,
        end_time: inputSnapshot.end_time,
      };

      expect(migratedEvent).toEqual(expectedEvent);
    });

    test("Unsupported schema", () => {
      const inputSnapshot: DashboardExecutionEventWithSchema = {
        // @ts-ignore
        schema_version: "20221010",
      };

      expect(() => stripSnapshotDataForExport(inputSnapshot)).toThrow(
        `Unsupported dashboard event schema ${inputSnapshot.schema_version}`
      );
    });
  });
});
