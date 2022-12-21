import {
  controlsUpdatedEventHandler,
  leafNodesUpdatedEventHandler,
  migrateDashboardExecutionCompleteSchema,
  migrateSnapshotDataToExecutionCompleteEvent,
} from "./dashboardEventHandlers";
import { DashboardActions, DashboardExecutionEventWithSchema } from "../types";
import { LATEST_EXECUTION_SCHEMA_VERSION } from "../constants/versions";

describe("dashboard event handlers", () => {
  describe("migrateSnapshotDataToExecutionCompleteEvent", () => {
    test("Schema 20220614 to 20220929", () => {
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
          },
          "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0":
            {
              name: "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0",
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

      const migratedEvent =
        migrateSnapshotDataToExecutionCompleteEvent(inputSnapshot);

      const expectedEvent = {
        action: DashboardActions.EXECUTION_COMPLETE,
        schema_version: LATEST_EXECUTION_SCHEMA_VERSION,
        snapshot: {
          schema_version: LATEST_EXECUTION_SCHEMA_VERSION,
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

    test("Schema 20220929 to 20220929", () => {
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
          },
          "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0":
            {
              name: "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0",
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

      const migratedEvent =
        migrateSnapshotDataToExecutionCompleteEvent(inputSnapshot);

      const expectedEvent = {
        action: DashboardActions.EXECUTION_COMPLETE,
        schema_version: LATEST_EXECUTION_SCHEMA_VERSION,
        snapshot: {
          schema_version: LATEST_EXECUTION_SCHEMA_VERSION,
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
      const inputSnapshot: DashboardExecutionEventWithSchema = {
        // @ts-ignore
        schema_version: "20221010",
      };

      expect(() =>
        migrateSnapshotDataToExecutionCompleteEvent(inputSnapshot)
      ).toThrow(
        `Unsupported dashboard event schema ${inputSnapshot.schema_version}`
      );
    });
  });

  describe("migrateDashboardExecutionCompleteSchema", () => {
    test("Schema 20220614 to 20220929", () => {
      const inputEvent: DashboardExecutionEventWithSchema = {
        action: "execution_complete",
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
          },
          "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0":
            {
              name: "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0",
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

      const migratedEvent = migrateDashboardExecutionCompleteSchema(inputEvent);

      const expectedEvent = {
        action: inputEvent.action,
        schema_version: LATEST_EXECUTION_SCHEMA_VERSION,
        execution_id: inputEvent.execution_id,
        snapshot: {
          schema_version: LATEST_EXECUTION_SCHEMA_VERSION,
          layout: inputEvent.layout,
          panels: inputEvent.panels,
          inputs: inputEvent.inputs,
          variables: inputEvent.variables,
          search_path: inputEvent.search_path,
          start_time: inputEvent.start_time,
          end_time: inputEvent.end_time,
        },
      };

      expect(migratedEvent).toEqual(expectedEvent);
    });

    test("Schema 20220929 to 20220929", () => {
      const inputEvent: DashboardExecutionEventWithSchema = {
        action: "execution_complete",
        schema_version: "20220929",
        execution_id: "0x140029247e0",
        snapshot: {
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
            },
            "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0":
              {
                name: "aws_insights.container.dashboard_aws_iam_user_dashboard_anonymous_container_0",
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

      const migratedEvent = migrateDashboardExecutionCompleteSchema(inputEvent);

      expect(migratedEvent).toEqual(inputEvent);
    });

    test("Unsupported schema", () => {
      const inputEvent: DashboardExecutionEventWithSchema = {
        // @ts-ignore
        schema_version: "20221010",
      };

      expect(() => migrateDashboardExecutionCompleteSchema(inputEvent)).toThrow(
        `Unsupported dashboard event schema ${inputEvent.schema_version}`
      );
    });
  });

  describe("controlsUpdatedEventHandler", () => {
    test("ignore complete events", () => {
      const state = { state: "complete" };
      expect(controlsUpdatedEventHandler(null, state)).toEqual(state);
    });

    test("no event controls", () => {
      const state = { state: "ready" };
      expect(controlsUpdatedEventHandler({ controls: null }, state)).toEqual(
        state
      );
    });

    test("empty event controls", () => {
      const state = { state: "ready" };
      expect(controlsUpdatedEventHandler({ controls: [] }, state)).toEqual(
        state
      );
    });

    test("control for different execution", () => {
      const state = {
        state: "ready",
        execution_id: "1",
        panelsMap: {},
        progress: 100,
      };
      expect(
        controlsUpdatedEventHandler(
          { controls: [{ execution_id: "2" }] },
          state
        )
      ).toEqual(state);
    });

    test("single control complete", () => {
      const state = {
        state: "ready",
        execution_id: "1",
        panelsMap: {
          control_a: {
            name: "control_a",
            panel_type: "control",
            status: "ready",
          },
          control_b: {
            name: "control_b",
            panel_type: "control",
            status: "ready",
          },
        },
        progress: 0,
      };
      const updatedControl = {
        name: "control_b",
        panel_type: "control",
        status: "complete",
      };
      expect(
        controlsUpdatedEventHandler(
          {
            controls: [
              {
                control: updatedControl,
                execution_id: "1",
              },
            ],
          },
          state
        )
      ).toEqual({
        ...state,
        panelsMap: {
          ...state.panelsMap,
          [updatedControl.name]: updatedControl,
        },
        progress: 50,
      });
    });

    test("single control error", () => {
      const state = {
        state: "ready",
        execution_id: "1",
        panelsMap: {
          control_a: {
            name: "control_a",
            panel_type: "control",
            status: "ready",
          },
          control_b: {
            name: "control_b",
            panel_type: "control",
            status: "ready",
          },
        },
        progress: 0,
      };
      const updatedControl = {
        name: "control_b",
        panel_type: "control",
        status: "error",
      };
      expect(
        controlsUpdatedEventHandler(
          {
            controls: [
              {
                control: updatedControl,
                execution_id: "1",
              },
            ],
          },
          state
        )
      ).toEqual({
        ...state,
        panelsMap: {
          ...state.panelsMap,
          [updatedControl.name]: updatedControl,
        },
        progress: 50,
      });
    });

    test("multiple controls", () => {
      const state = {
        state: "ready",
        execution_id: "1",
        panelsMap: {
          control_a: {
            name: "control_a",
            panel_type: "control",
            status: "ready",
          },
          control_b: {
            name: "control_b",
            panel_type: "control",
            status: "ready",
          },
          control_c: {
            name: "control_c",
            panel_type: "control",
            status: "ready",
          },
          control_d: {
            name: "control_d",
            panel_type: "control",
            status: "ready",
          },
        },
        progress: 0,
      };
      const updatedControl1 = {
        name: "control_a",
        panel_type: "control",
        status: "complete",
      };
      const updatedControl2 = {
        name: "control_b",
        panel_type: "control",
        status: "error",
      };
      const updatedControl3 = {
        name: "control_d",
        panel_type: "control",
        status: "complete",
      };
      expect(
        controlsUpdatedEventHandler(
          {
            controls: [
              {
                control: updatedControl1,
                execution_id: "1",
              },
              {
                control: updatedControl2,
                execution_id: "1",
              },
              {
                control: updatedControl3,
                execution_id: "1",
              },
            ],
          },
          state
        )
      ).toEqual({
        ...state,
        panelsMap: {
          ...state.panelsMap,
          [updatedControl1.name]: updatedControl1,
          [updatedControl2.name]: updatedControl2,
          [updatedControl3.name]: updatedControl3,
        },
        progress: 75,
      });
    });
  });

  describe("leafNodesUpdatedEventHandler", () => {
    test("ignore complete events", () => {
      const state = { state: "complete" };
      expect(leafNodesUpdatedEventHandler(null, state)).toEqual(state);
    });

    test("no event nodes", () => {
      const state = { state: "ready" };
      expect(leafNodesUpdatedEventHandler({ nodes: null }, state)).toEqual(
        state
      );
    });

    test("empty event nodes", () => {
      const state = { state: "ready" };
      expect(leafNodesUpdatedEventHandler({ nodes: [] }, state)).toEqual(state);
    });

    test("node for different execution", () => {
      const state = {
        state: "ready",
        execution_id: "1",
        panelsLog: {},
        panelsMap: {},
        progress: 100,
      };
      expect(
        leafNodesUpdatedEventHandler({ nodes: [{ execution_id: "2" }] }, state)
      ).toEqual(state);
    });

    test("single node blocked", () => {
      const readyAt = new Date();
      const blockedAt = new Date(readyAt);
      blockedAt.setSeconds(readyAt.getSeconds() + 1);
      const state = {
        state: "ready",
        execution_id: "1",
        panelsLog: {
          panel_a: [
            {
              error: null,
              status: "ready",
              timestamp: readyAt.valueOf(),
              title: "panel_a",
            },
          ],
        },
        panelsMap: { panel_a: { name: "panel_a", sql: "", status: "ready" } },
        progress: 0,
      };
      const updatedDashboardNode = {
        name: "panel_a",
        panel_type: "node",
        sql: "",
        status: "blocked",
        error: null,
      };
      expect(
        leafNodesUpdatedEventHandler(
          {
            nodes: [
              {
                dashboard_node: updatedDashboardNode,
                execution_id: "1",
                timestamp: blockedAt.valueOf(),
              },
            ],
          },
          state
        )
      ).toEqual({
        ...state,
        panelsLog: {
          ...state.panelsLog,
          panel_a: [
            ...state.panelsLog.panel_a,
            {
              error: null,
              status: "blocked",
              timestamp: blockedAt.valueOf(),
              title: "panel_a",
            },
          ],
        },
        panelsMap: {
          ...state.panelsMap,
          [updatedDashboardNode.name]: updatedDashboardNode,
        },
        progress: 0,
      });
    });

    test("single node running", () => {
      const readyAt = new Date();
      const runningAt = new Date(readyAt);
      runningAt.setSeconds(readyAt.getSeconds() + 1);
      const state = {
        state: "ready",
        execution_id: "1",
        panelsLog: {
          panel_a: [
            {
              error: null,
              status: "ready",
              timestamp: readyAt.valueOf(),
              title: "panel_a",
            },
          ],
        },
        panelsMap: { panel_a: { name: "panel_a", sql: "", status: "ready" } },
        progress: 0,
      };
      const updatedDashboardNode = {
        name: "panel_a",
        panel_type: "node",
        sql: "",
        status: "running",
        error: null,
      };
      expect(
        leafNodesUpdatedEventHandler(
          {
            nodes: [
              {
                dashboard_node: updatedDashboardNode,
                execution_id: "1",
                timestamp: runningAt.valueOf(),
              },
            ],
          },
          state
        )
      ).toEqual({
        ...state,
        panelsLog: {
          ...state.panelsLog,
          panel_a: [
            ...state.panelsLog.panel_a,
            {
              error: null,
              status: "running",
              timestamp: runningAt.valueOf(),
              title: "panel_a",
            },
          ],
        },
        panelsMap: {
          ...state.panelsMap,
          [updatedDashboardNode.name]: updatedDashboardNode,
        },
        progress: 0,
      });
    });

    test("single node error", () => {
      const readyAt = new Date();
      const erroredAt = new Date(readyAt);
      erroredAt.setSeconds(readyAt.getSeconds() + 1);
      const state = {
        state: "ready",
        execution_id: "1",
        panelsLog: {
          panel_a: [
            {
              error: null,
              status: "ready",
              timestamp: readyAt.valueOf(),
              title: "panel_a",
            },
          ],
        },
        panelsMap: { panel_a: { name: "panel_a", sql: "", status: "ready" } },
        progress: 0,
      };
      const updatedDashboardNode = {
        name: "panel_a",
        panel_type: "node",
        sql: "",
        status: "error",
        error: "BOOM!",
      };
      expect(
        leafNodesUpdatedEventHandler(
          {
            nodes: [
              {
                dashboard_node: updatedDashboardNode,
                execution_id: "1",
                timestamp: erroredAt.valueOf(),
              },
            ],
          },
          state
        )
      ).toEqual({
        ...state,
        panelsLog: {
          ...state.panelsLog,
          panel_a: [
            ...state.panelsLog.panel_a,
            {
              error: "BOOM!",
              status: "error",
              timestamp: erroredAt.valueOf(),
              title: "panel_a",
            },
          ],
        },
        panelsMap: {
          ...state.panelsMap,
          [updatedDashboardNode.name]: updatedDashboardNode,
        },
        progress: 100,
      });
    });

    test("single node complete", () => {
      const readyAt = new Date();
      const completeAt = new Date(readyAt);
      completeAt.setSeconds(readyAt.getSeconds() + 1);
      const state = {
        state: "running",
        execution_id: "1",
        panelsLog: {
          panel_a: [
            {
              error: null,
              status: "running",
              timestamp: readyAt.valueOf(),
              title: "panel_a",
            },
          ],
        },
        panelsMap: { panel_a: { name: "panel_a", sql: "", status: "ready" } },
        progress: 0,
      };
      const updatedDashboardNode = {
        name: "panel_a",
        panel_type: "node",
        sql: "",
        status: "complete",
        error: null,
      };
      expect(
        leafNodesUpdatedEventHandler(
          {
            nodes: [
              {
                dashboard_node: updatedDashboardNode,
                execution_id: "1",
                timestamp: completeAt.valueOf(),
              },
            ],
          },
          state
        )
      ).toEqual({
        ...state,
        panelsLog: {
          ...state.panelsLog,
          panel_a: [
            ...state.panelsLog.panel_a,
            {
              error: null,
              executionTime: 1000,
              status: "complete",
              timestamp: completeAt.valueOf(),
              title: "panel_a",
            },
          ],
        },
        panelsMap: {
          ...state.panelsMap,
          [updatedDashboardNode.name]: updatedDashboardNode,
        },
        progress: 100,
      });
    });

    test("multiple node complete", () => {
      const readyAt = new Date();
      const panelACompleteAt = new Date(readyAt);
      panelACompleteAt.setSeconds(readyAt.getSeconds() + 1);
      const panelBCompleteAt = new Date(readyAt);
      panelBCompleteAt.setSeconds(readyAt.getSeconds() + 2);
      const state = {
        state: "ready",
        execution_id: "1",
        panelsLog: {
          panel_a: [
            {
              error: null,
              status: "ready",
              timestamp: readyAt.valueOf(),
              title: "panel_a",
            },
          ],
          panel_b: [
            {
              error: null,
              status: "ready",
              timestamp: readyAt.valueOf(),
              title: "panel_b",
            },
          ],
          panel_c: [
            {
              error: null,
              status: "ready",
              timestamp: readyAt.valueOf(),
              title: "panel_c",
            },
          ],
          panel_d: [
            {
              error: null,
              status: "ready",
              timestamp: readyAt.valueOf(),
              title: "panel_d",
            },
          ],
        },
        panelsMap: {
          panel_a: { name: "panel_a", sql: "", status: "ready" },
          panel_b: { name: "panel_b", sql: "", status: "ready" },
          panel_c: { name: "panel_c", sql: "", status: "ready" },
          panel_d: { name: "panel_d", sql: "", status: "ready" },
        },
        progress: 0,
      };
      const updatedDashboardNode1 = {
        name: "panel_a",
        panel_type: "node",
        sql: "",
        status: "complete",
        error: null,
      };
      const updatedDashboardNode2 = {
        name: "panel_b",
        panel_type: "edge",
        sql: "",
        status: "complete",
        error: null,
      };
      expect(
        leafNodesUpdatedEventHandler(
          {
            nodes: [
              {
                dashboard_node: updatedDashboardNode1,
                execution_id: "1",
                timestamp: panelACompleteAt.valueOf(),
              },
              {
                dashboard_node: updatedDashboardNode2,
                execution_id: "1",
                timestamp: panelBCompleteAt.valueOf(),
              },
            ],
          },
          state
        )
      ).toEqual({
        ...state,
        panelsLog: {
          ...state.panelsLog,
          panel_a: [
            ...state.panelsLog.panel_a,
            {
              error: null,
              status: "complete",
              timestamp: panelACompleteAt.valueOf(),
              title: "panel_a",
            },
          ],
          panel_b: [
            ...state.panelsLog.panel_b,
            {
              error: null,
              status: "complete",
              timestamp: panelBCompleteAt.valueOf(),
              title: "panel_b",
            },
          ],
        },
        panelsMap: {
          ...state.panelsMap,
          [updatedDashboardNode1.name]: updatedDashboardNode1,
          [updatedDashboardNode2.name]: updatedDashboardNode2,
        },
        progress: 50,
      });
    });
  });
});
