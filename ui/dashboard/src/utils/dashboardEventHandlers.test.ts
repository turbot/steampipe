import {
  controlsUpdatedEventHandler,
  leafNodesUpdatedEventHandler,
} from "./dashboardEventHandlers";
import { EXECUTION_SCHEMA_VERSION_20221222 } from "../constants/versions";

describe("dashboard event handlers", () => {
  describe("controlsUpdatedEventHandler", () => {
    test("ignore complete events", () => {
      const state = { state: "complete" };
      expect(controlsUpdatedEventHandler(null, state)).toEqual(state);
    });

    test("no event controls", () => {
      const state = { state: "running" };
      expect(controlsUpdatedEventHandler({ controls: null }, state)).toEqual(
        state
      );
    });

    test("empty event controls", () => {
      const state = { state: "running" };
      expect(controlsUpdatedEventHandler({ controls: [] }, state)).toEqual(
        state
      );
    });

    test("control for different execution", () => {
      const state = {
        state: "running",
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
        state: "running",
        execution_id: "1",
        panelsMap: {
          control_a: {
            name: "control_a",
            panel_type: "control",
            status: "running",
          },
          control_b: {
            name: "control_b",
            panel_type: "control",
            status: "running",
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
        state: "running",
        execution_id: "1",
        panelsMap: {
          control_a: {
            name: "control_a",
            panel_type: "control",
            status: "running",
          },
          control_b: {
            name: "control_b",
            panel_type: "control",
            status: "running",
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
        state: "running",
        execution_id: "1",
        panelsMap: {
          control_a: {
            name: "control_a",
            panel_type: "control",
            status: "running",
          },
          control_b: {
            name: "control_b",
            panel_type: "control",
            status: "running",
          },
          control_c: {
            name: "control_c",
            panel_type: "control",
            status: "running",
          },
          control_d: {
            name: "control_d",
            panel_type: "control",
            status: "running",
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
      expect(
        leafNodesUpdatedEventHandler(
          null,
          EXECUTION_SCHEMA_VERSION_20221222,
          state
        )
      ).toEqual(state);
    });

    test("no event nodes", () => {
      const state = { state: "running" };
      expect(
        leafNodesUpdatedEventHandler(
          { nodes: null },
          EXECUTION_SCHEMA_VERSION_20221222,
          state
        )
      ).toEqual(state);
    });

    test("empty event nodes", () => {
      const state = { state: "running" };
      expect(
        leafNodesUpdatedEventHandler(
          { nodes: [] },
          EXECUTION_SCHEMA_VERSION_20221222,
          state
        )
      ).toEqual(state);
    });

    test("node for different execution", () => {
      const state = {
        state: "running",
        execution_id: "1",
        panelsLog: {},
        panelsMap: {},
        progress: 100,
      };
      expect(
        leafNodesUpdatedEventHandler(
          { nodes: [{ execution_id: "2" }] },
          EXECUTION_SCHEMA_VERSION_20221222,
          state
        )
      ).toEqual(state);
    });

    test("single node blocked", () => {
      const readyAt = new Date();
      const blockedAt = new Date(readyAt);
      blockedAt.setSeconds(readyAt.getSeconds() + 1);
      const state = {
        state: "running",
        execution_id: "1",
        panelsLog: {
          panel_a: [
            {
              error: null,
              status: "running",
              timestamp: readyAt.toString(),
              title: "panel_a",
            },
          ],
        },
        panelsMap: { panel_a: { name: "panel_a", sql: "", status: "running" } },
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
                timestamp: blockedAt.toString(),
              },
            ],
          },
          EXECUTION_SCHEMA_VERSION_20221222,
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
              timestamp: blockedAt.toString(),
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
        state: "initialized",
        execution_id: "1",
        panelsLog: {
          panel_a: [
            {
              error: null,
              status: "initialized",
              timestamp: readyAt.toString(),
              title: "panel_a",
            },
          ],
        },
        panelsMap: {
          panel_a: { name: "panel_a", sql: "", status: "initialized" },
        },
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
                timestamp: runningAt.toString(),
              },
            ],
          },
          EXECUTION_SCHEMA_VERSION_20221222,
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
              timestamp: runningAt.toString(),
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
        state: "running",
        execution_id: "1",
        panelsLog: {
          panel_a: [
            {
              error: null,
              status: "running",
              timestamp: readyAt.toString(),
              title: "panel_a",
            },
          ],
        },
        panelsMap: { panel_a: { name: "panel_a", sql: "", status: "running" } },
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
                timestamp: erroredAt.toString(),
              },
            ],
          },
          EXECUTION_SCHEMA_VERSION_20221222,
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
              timestamp: erroredAt.toString(),
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
              timestamp: readyAt.toString(),
              title: "panel_a",
            },
          ],
        },
        panelsMap: { panel_a: { name: "panel_a", sql: "", status: "running" } },
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
                timestamp: completeAt.toString(),
              },
            ],
          },
          EXECUTION_SCHEMA_VERSION_20221222,
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
              timestamp: completeAt.toString(),
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
        state: "running",
        execution_id: "1",
        panelsLog: {
          panel_a: [
            {
              error: null,
              status: "running",
              timestamp: readyAt.toString(),
              title: "panel_a",
            },
          ],
          panel_b: [
            {
              error: null,
              status: "running",
              timestamp: readyAt.toString(),
              title: "panel_b",
            },
          ],
          panel_c: [
            {
              error: null,
              status: "running",
              timestamp: readyAt.toString(),
              title: "panel_c",
            },
          ],
          panel_d: [
            {
              error: null,
              status: "running",
              timestamp: readyAt.toString(),
              title: "panel_d",
            },
          ],
        },
        panelsMap: {
          panel_a: { name: "panel_a", sql: "", status: "running" },
          panel_b: { name: "panel_b", sql: "", status: "running" },
          panel_c: { name: "panel_c", sql: "", status: "running" },
          panel_d: { name: "panel_d", sql: "", status: "running" },
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
                timestamp: panelACompleteAt.toString(),
              },
              {
                dashboard_node: updatedDashboardNode2,
                execution_id: "1",
                timestamp: panelBCompleteAt.toString(),
              },
            ],
          },
          EXECUTION_SCHEMA_VERSION_20221222,
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
              timestamp: panelACompleteAt.toString(),
              title: "panel_a",
            },
          ],
          panel_b: [
            ...state.panelsLog.panel_b,
            {
              error: null,
              executionTime: 2000,
              status: "complete",
              timestamp: panelBCompleteAt.toString(),
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
