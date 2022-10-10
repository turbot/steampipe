import {
  controlsUpdatedEventHandler,
  leafNodesCompleteEventHandler,
} from "./dashboardEventHandlers";

describe("dashboard event handlers", () => {
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

  describe("leafNodesCompleteEventHandler", () => {
    test("ignore complete events", () => {
      const state = { state: "complete" };
      expect(leafNodesCompleteEventHandler(null, state)).toEqual(state);
    });

    test("no event nodes", () => {
      const state = { state: "ready" };
      expect(leafNodesCompleteEventHandler({ nodes: null }, state)).toEqual(
        state
      );
    });

    test("empty event nodes", () => {
      const state = { state: "ready" };
      expect(leafNodesCompleteEventHandler({ nodes: [] }, state)).toEqual(
        state
      );
    });

    test("node for different execution", () => {
      const state = {
        state: "ready",
        execution_id: "1",
        panelsMap: {},
        progress: 100,
      };
      expect(
        leafNodesCompleteEventHandler({ nodes: [{ execution_id: "2" }] }, state)
      ).toEqual(state);
    });

    test("single node complete", () => {
      const state = {
        state: "ready",
        execution_id: "1",
        panelsMap: { panel_a: { name: "panel_a", sql: "", status: "ready" } },
        progress: 0,
      };
      const updatedDashboardNode = {
        name: "panel_a",
        sql: "",
        status: "complete",
      };
      expect(
        leafNodesCompleteEventHandler(
          {
            nodes: [
              {
                dashboard_node: updatedDashboardNode,
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
          [updatedDashboardNode.name]: updatedDashboardNode,
        },
        progress: 100,
      });
    });

    test("multiple node complete", () => {
      const state = {
        state: "ready",
        execution_id: "1",
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
        sql: "",
        status: "complete",
      };
      const updatedDashboardNode2 = {
        name: "panel_b",
        sql: "",
        status: "complete",
      };
      expect(
        leafNodesCompleteEventHandler(
          {
            nodes: [
              {
                dashboard_node: updatedDashboardNode1,
                execution_id: "1",
              },
              {
                dashboard_node: updatedDashboardNode2,
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
          [updatedDashboardNode1.name]: updatedDashboardNode1,
          [updatedDashboardNode2.name]: updatedDashboardNode2,
        },
        progress: 50,
      });
    });
  });
});
