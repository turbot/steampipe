import { useReducer } from "react";
import sortBy from "lodash/sortBy";
import {
  AvailableDashboard,
  AvailableDashboardsDictionary,
  DashboardActions,
  DashboardsCollection,
} from "../../types/dashboard";

const buildDashboards = (
  dashboards: AvailableDashboardsDictionary,
  benchmarks: AvailableDashboardsDictionary
): DashboardsCollection => {
  const dashboardsMap = {};
  const builtDashboards: AvailableDashboard[] = [];

  for (const [, dashboard] of Object.entries(dashboards)) {
    const builtDashboard: AvailableDashboard = {
      title: dashboard.title,
      full_name: dashboard.full_name,
      short_name: dashboard.short_name,
      type: "dashboard",
      tags: dashboard.tags,
      mod_full_name: dashboard.mod_full_name,
      is_top_level: true,
    };
    dashboardsMap[builtDashboard.full_name] = builtDashboard;
    builtDashboards.push(builtDashboard);
  }

  for (const [, benchmark] of Object.entries(benchmarks)) {
    const builtBenchmark: AvailableDashboard = {
      title: benchmark.title,
      full_name: benchmark.full_name,
      short_name: benchmark.short_name,
      type: "benchmark",
      tags: benchmark.tags,
      mod_full_name: benchmark.mod_full_name,
      is_top_level: benchmark.is_top_level,
      trunks: benchmark.trunks,
      children: benchmark.children,
    };
    dashboardsMap[builtBenchmark.full_name] = builtBenchmark;
    builtDashboards.push(builtBenchmark);
  }

  return {
    dashboards: sortBy(builtDashboards, [
      (dashboard) =>
        dashboard.title
          ? dashboard.title.toLowerCase()
          : dashboard.full_name.toLowerCase(),
    ]),
    dashboardsMap,
  };
};

const dashboardReducer = (state, action) => {
  switch (action.type) {
    case DashboardActions.DASHBOARD_METADATA:
      return {
        ...state,
        metadata: action.metadata,
      };
    case DashboardActions.AVAILABLE_DASHBOARDS:
      const { dashboards, dashboardsMap } = buildDashboards(
        action.dashboards,
        action.benchmarks
      );
      return {
        ...state,
        dashboards,
        dashboardsMap,
      };
    default:
      return state;
  }
};

const getInitialDashboardState = () => {};

const useDashboardState = () => {
  const [dashboardState, dispatch] = useReducer(
    dashboardReducer,
    getInitialDashboardState
  );
  return {
    state: dashboardState,
    dispatch,
  };
};

export default useDashboardState;
