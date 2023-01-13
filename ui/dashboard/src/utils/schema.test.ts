import {
  DashboardActions,
  DashboardExecutionEventWithSchema,
  DashboardSnapshot,
} from "../types";
import {
  EXECUTION_COMPLETE_SCHEMA_VERSION_LATEST,
  EXECUTION_SCHEMA_VERSION_20220614,
  EXECUTION_SCHEMA_VERSION_20220929,
  EXECUTION_SCHEMA_VERSION_20221222,
} from "../constants/versions";
import {
  ExecutionCompleteSchemaMigrator,
  ExecutionStartedSchemaMigrator,
  SnapshotDataToExecutionCompleteSchemaMigrator,
} from "./schema";

describe("schema", () => {
  describe("execution_started schema migrations", () => {
    test("Schema 20220614 to 20221222", () => {
      const inputEvent: DashboardExecutionEventWithSchema = {
        action: "execution_started",
        schema_version: EXECUTION_SCHEMA_VERSION_20220614,
        execution_id: "0x140029247e0",
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
            // @ts-ignore
            status: "ready",
          },
          "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0":
            {
              name: "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0",
              // @ts-ignore
              status: "ready",
            },
        },
        inputs: {
          "input.foo": "bar",
        },
        variables: {
          foo: "bar",
        },
      };

      const eventMigrator = new ExecutionStartedSchemaMigrator();
      const migratedEvent = eventMigrator.toLatest(inputEvent);

      const expectedPanels = {
        "aws_insights.dashboard.aws_iam_user_dashboard": {
          name: "aws_insights.dashboard.aws_iam_user_dashboard",
          status: "running",
        },
        "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0":
          {
            name: "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0",
            status: "running",
          },
      };

      const expectedEvent = {
        action: inputEvent.action,
        schema_version: EXECUTION_COMPLETE_SCHEMA_VERSION_LATEST,
        execution_id: inputEvent.execution_id,
        layout: inputEvent.layout,
        panels: expectedPanels,
        inputs: inputEvent.inputs,
        variables: inputEvent.variables,
      };

      expect(migratedEvent).toEqual(expect.objectContaining(expectedEvent));
    });

    test("Unsupported schema", () => {
      const inputEvent: DashboardExecutionEventWithSchema = {
        // @ts-ignore
        schema_version: "20221010",
      };

      const eventMigrator = new ExecutionCompleteSchemaMigrator();

      expect(() => eventMigrator.toLatest(inputEvent)).toThrow(
        `Unsupported dashboard event schema ${inputEvent.schema_version}`
      );
    });
  });

  describe("execution_complete schema migrations", () => {
    test("Schema 20220614 to 20221222", () => {
      const inputEvent: DashboardExecutionEventWithSchema = {
        action: "execution_complete",
        schema_version: EXECUTION_SCHEMA_VERSION_20220614,
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
            // @ts-ignore
            status: "ready",
          },
          "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0":
            {
              name: "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0",
              // @ts-ignore
              status: "ready",
            },
        },
        inputs: {
          "input.foo": "bar",
        },
        variables: {
          foo: "bar",
        },
        search_path: ["some_schema"],
        start_time: "2022-10-27T14:43:57.79514+01:00",
        end_time: "2022-10-27T14:43:58.045925+01:00",
      };

      const eventMigrator = new ExecutionCompleteSchemaMigrator();
      const migratedEvent = eventMigrator.toLatest(inputEvent);

      const expectedPanels = {
        "aws_insights.dashboard.aws_iam_user_dashboard": {
          name: "aws_insights.dashboard.aws_iam_user_dashboard",
          status: "running",
        },
        "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0":
          {
            name: "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0",
            status: "running",
          },
      };

      const expectedEvent = {
        action: inputEvent.action,
        schema_version: EXECUTION_COMPLETE_SCHEMA_VERSION_LATEST,
        execution_id: inputEvent.execution_id,
        snapshot: {
          schema_version: EXECUTION_COMPLETE_SCHEMA_VERSION_LATEST,
          layout: inputEvent.layout,
          panels: expectedPanels,
          inputs: inputEvent.inputs,
          variables: inputEvent.variables,
          search_path: inputEvent.search_path,
          end_time: inputEvent.end_time,
          start_time: inputEvent.start_time,
        },
      };

      expect(migratedEvent).toEqual(expectedEvent);
    });

    test("Schema 20220929 to 20221222", () => {
      const inputEvent: DashboardExecutionEventWithSchema = {
        action: "execution_complete",
        schema_version: EXECUTION_SCHEMA_VERSION_20220929,
        execution_id: "0x140029247e0",
        snapshot: {
          schema_version: EXECUTION_SCHEMA_VERSION_20220929,
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
              // @ts-ignore
              status: "ready",
            },
            "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0":
              {
                name: "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0",
                // @ts-ignore
                status: "ready",
              },
          },
          inputs: {
            "input.foo": "bar",
          },
          variables: {
            foo: "bar",
          },
          search_path: ["some_schema"],
          start_time: "2022-10-27T14:43:57.79514+01:00",
          end_time: "2022-10-27T14:43:58.045925+01:00",
        },
      };

      const eventMigrator = new ExecutionCompleteSchemaMigrator();
      const migratedEvent = eventMigrator.toLatest(inputEvent);

      const expectedPanels = {
        "aws_insights.dashboard.aws_iam_user_dashboard": {
          name: "aws_insights.dashboard.aws_iam_user_dashboard",
          status: "running",
        },
        "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0":
          {
            name: "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0",
            status: "running",
          },
      };

      const expectedEvent = {
        action: inputEvent.action,
        schema_version: EXECUTION_COMPLETE_SCHEMA_VERSION_LATEST,
        execution_id: inputEvent.execution_id,
        snapshot: {
          schema_version: EXECUTION_COMPLETE_SCHEMA_VERSION_LATEST,
          layout: inputEvent.snapshot.layout,
          panels: expectedPanels,
          inputs: inputEvent.snapshot.inputs,
          variables: inputEvent.snapshot.variables,
          search_path: inputEvent.snapshot.search_path,
          start_time: inputEvent.snapshot.start_time,
          end_time: inputEvent.snapshot.end_time,
        },
      };

      expect(migratedEvent).toEqual(expectedEvent);
    });

    test("Schema 20221222 to 20221222", () => {
      const inputEvent: DashboardExecutionEventWithSchema = {
        action: "execution_complete",
        schema_version: EXECUTION_SCHEMA_VERSION_20221222,
        execution_id: "0x140029247e0",
        snapshot: {
          schema_version: EXECUTION_SCHEMA_VERSION_20221222,
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
              status: "blocked",
            },
            "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0":
              {
                name: "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0",
                // @ts-ignore
                status: "blocked",
              },
          },
          inputs: {
            "input.foo": "bar",
          },
          variables: {
            foo: "bar",
          },
          search_path: ["some_schema"],
          start_time: "2022-10-27T14:43:57.79514+01:00",
          end_time: "2022-10-27T14:43:58.045925+01:00",
        },
      };

      const eventMigrator = new ExecutionCompleteSchemaMigrator();
      const migratedEvent = eventMigrator.toLatest(inputEvent);

      expect(migratedEvent).toEqual(inputEvent);
    });

    test("Unsupported schema", () => {
      const inputEvent: DashboardExecutionEventWithSchema = {
        // @ts-ignore
        schema_version: "20221010",
      };

      const eventMigrator = new ExecutionCompleteSchemaMigrator();

      expect(() => eventMigrator.toLatest(inputEvent)).toThrow(
        `Unsupported dashboard event schema ${inputEvent.schema_version}`
      );
    });
  });

  describe("snapshot data to execution_complete event", () => {
    test("Schema 20220614 to 20221222", () => {
      const inputSnapshot: DashboardSnapshot = {
        schema_version: EXECUTION_SCHEMA_VERSION_20220614,
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
            panel_type: "dashboard",
            dashboard: "aws_insights.dashboard.aws_iam_user_dashboard",
            // @ts-ignore
            status: "ready",
          },
          "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0":
            {
              name: "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0",
              panel_type: "container",
              dashboard: "aws_insights.dashboard.aws_iam_user_dashboard",
              // @ts-ignore
              status: "ready",
            },
        },
        inputs: {
          "input.foo": "bar",
        },
        variables: {
          foo: "bar",
        },
        search_path: ["some_schema"],
        start_time: "2022-10-27T14:43:57.79514+01:00",
        end_time: "2022-10-27T14:43:58.045925+01:00",
      };

      const eventMigrator = new SnapshotDataToExecutionCompleteSchemaMigrator();
      const migratedEvent = eventMigrator.toLatest(inputSnapshot);

      const expectedPanels = {
        "aws_insights.dashboard.aws_iam_user_dashboard": {
          name: "aws_insights.dashboard.aws_iam_user_dashboard",
          panel_type: "dashboard",
          dashboard: "aws_insights.dashboard.aws_iam_user_dashboard",
          status: "running",
        },
        "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0":
          {
            name: "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0",
            panel_type: "container",
            dashboard: "aws_insights.dashboard.aws_iam_user_dashboard",
            status: "running",
          },
      };

      const expectedEvent = {
        action: DashboardActions.EXECUTION_COMPLETE,
        execution_id: "",
        schema_version: EXECUTION_COMPLETE_SCHEMA_VERSION_LATEST,
        snapshot: {
          schema_version: EXECUTION_COMPLETE_SCHEMA_VERSION_LATEST,
          layout: inputSnapshot.layout,
          panels: expectedPanels,
          inputs: inputSnapshot.inputs,
          variables: inputSnapshot.variables,
          search_path: inputSnapshot.search_path,
          start_time: inputSnapshot.start_time,
          end_time: inputSnapshot.end_time,
        },
      };

      expect(migratedEvent).toEqual(expectedEvent);
    });

    test("Schema 20220929 to 20221222", () => {
      const inputSnapshot: DashboardSnapshot = {
        schema_version: EXECUTION_SCHEMA_VERSION_20220929,
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
            panel_type: "dashboard",
            dashboard: "aws_insights.dashboard.aws_iam_user_dashboard",
            // @ts-ignore
            status: "ready",
          },
          "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0":
            {
              name: "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0",
              panel_type: "container",
              dashboard: "aws_insights.dashboard.aws_iam_user_dashboard",
              // @ts-ignore
              status: "ready",
            },
        },
        inputs: {
          "input.foo": "bar",
        },
        variables: {
          foo: "bar",
        },
        search_path: ["some_schema"],
        start_time: "2022-10-27T14:43:57.79514+01:00",
        end_time: "2022-10-27T14:43:58.045925+01:00",
      };

      const eventMigrator = new SnapshotDataToExecutionCompleteSchemaMigrator();
      const migratedEvent = eventMigrator.toLatest(inputSnapshot);

      const expectedPanels = {
        "aws_insights.dashboard.aws_iam_user_dashboard": {
          name: "aws_insights.dashboard.aws_iam_user_dashboard",
          panel_type: "dashboard",
          dashboard: "aws_insights.dashboard.aws_iam_user_dashboard",
          status: "running",
        },
        "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0":
          {
            name: "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0",
            panel_type: "container",
            dashboard: "aws_insights.dashboard.aws_iam_user_dashboard",
            status: "running",
          },
      };

      const expectedEvent = {
        action: DashboardActions.EXECUTION_COMPLETE,
        execution_id: "",
        schema_version: EXECUTION_COMPLETE_SCHEMA_VERSION_LATEST,
        snapshot: {
          schema_version: EXECUTION_COMPLETE_SCHEMA_VERSION_LATEST,
          layout: inputSnapshot.layout,
          panels: expectedPanels,
          inputs: inputSnapshot.inputs,
          variables: inputSnapshot.variables,
          search_path: inputSnapshot.search_path,
          start_time: inputSnapshot.start_time,
          end_time: inputSnapshot.end_time,
        },
      };

      expect(migratedEvent).toEqual(expectedEvent);
    });

    test("Schema 20221222 to 20221222", () => {
      const inputSnapshot: DashboardSnapshot = {
        schema_version: EXECUTION_SCHEMA_VERSION_20221222,
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
            panel_type: "dashboard",
            dashboard: "aws_insights.dashboard.aws_iam_user_dashboard",
            status: "blocked",
          },
          "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0":
            {
              name: "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0",
              panel_type: "container",
              dashboard: "aws_insights.dashboard.aws_iam_user_dashboard",
              status: "blocked",
            },
        },
        inputs: {
          "input.foo": "bar",
        },
        variables: {
          foo: "bar",
        },
        search_path: ["some_schema"],
        start_time: "2022-10-27T14:43:57.79514+01:00",
        end_time: "2022-10-27T14:43:58.045925+01:00",
      };

      const eventMigrator = new SnapshotDataToExecutionCompleteSchemaMigrator();
      const migratedEvent = eventMigrator.toLatest(inputSnapshot);

      const expectedEvent = {
        action: DashboardActions.EXECUTION_COMPLETE,
        execution_id: "",
        schema_version: EXECUTION_COMPLETE_SCHEMA_VERSION_LATEST,
        snapshot: {
          schema_version: EXECUTION_COMPLETE_SCHEMA_VERSION_LATEST,
          layout: inputSnapshot.layout,
          panels: inputSnapshot.panels,
          inputs: inputSnapshot.inputs,
          variables: inputSnapshot.variables,
          search_path: inputSnapshot.search_path,
          start_time: inputSnapshot.start_time,
          end_time: inputSnapshot.end_time,
        },
      };

      expect(migratedEvent).toEqual(expectedEvent);
    });

    test("Unsupported schema", () => {
      const inputSnapshot: DashboardSnapshot = {
        // @ts-ignore
        schema_version: "20221010",
      };

      const eventMigrator = new SnapshotDataToExecutionCompleteSchemaMigrator();

      expect(() => eventMigrator.toLatest(inputSnapshot)).toThrow(
        `Unsupported dashboard event schema ${inputSnapshot.schema_version}`
      );
    });
  });
});
