import get from "lodash/get";
import paths from "deepdash/paths";
import set from "lodash/set";
import sortBy from "lodash/sortBy";
import {
  AvailableDashboard,
  AvailableDashboardsDictionary,
  DashboardDefinition,
  DashboardRunState,
  DashboardsCollection,
  PanelDefinition,
  PanelLog,
  PanelsLog,
  PanelsMap,
  SQLDataMap,
} from "../types";
import { KeyValueStringPairs } from "../components/dashboards/common/types";

const addDataToPanels = (
  panels: PanelsMap,
  sqlDataMap: SQLDataMap
): PanelsMap => {
  const sqlPaths = paths(panels, { leavesOnly: true }).filter((path) =>
    path.toString().endsWith(".sql")
  );
  for (const sqlPath of sqlPaths) {
    const panelPath = sqlPath.toString().substring(0, sqlPath.indexOf(".sql"));
    const panel = get(panels, panelPath);
    const args = panel.args;
    const sql = panel.sql;
    if (!sql) {
      continue;
    }
    const key = getSqlDataMapKey(sql, args);
    // @ts-ignore
    const data = sqlDataMap[key];
    if (!data) {
      continue;
    }

    const dataPath = `${panelPath}.data`;
    // We don't want to retain panel data for inputs as it causes issues with selection
    // of incorrect values for select controls without placeholders
    if (panel && panel.panel_type !== "input") {
      set(panels, dataPath, data);
    }
  }
  return { ...panels };
};

const addPanelLog = (
  panelsLog: PanelsLog,
  panelName: string,
  panelLog: PanelLog
) => {
  const newPanelsLog = { ...panelsLog };
  const newPanelLog = [...(newPanelsLog[panelName] || [])];
  newPanelLog.push(panelLog);
  newPanelsLog[panelName] = newPanelLog;
  return newPanelsLog;
};

const buildDashboards = (
  dashboards: AvailableDashboardsDictionary,
  benchmarks: AvailableDashboardsDictionary,
  snapshots: KeyValueStringPairs
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

  for (const snapshot of Object.keys(snapshots || {})) {
    const builtSnapshot: AvailableDashboard = {
      title: snapshot,
      full_name: snapshot,
      short_name: snapshot,
      type: "snapshot",
      tags: {},
      is_top_level: true,
    };
    dashboardsMap[builtSnapshot.full_name] = builtSnapshot;
    builtDashboards.push(builtSnapshot);
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

const buildPanelLog = (panel: PanelDefinition, timestamp: number): PanelLog => {
  return {
    error: panel.status === "error" ? panel.error : null,
    status: panel.status as DashboardRunState,
    timestamp,
  };
};

const buildPanelsLog = (panels: PanelsMap, timestamp: number) => {
  const panelsLog: PanelsLog = {};
  for (const [name, panel] of Object.entries(panels || {})) {
    panelsLog[name] = [buildPanelLog(panel, timestamp)];
  }
  return panelsLog;
};

const updatePanelsLogFromCompletedPanels = (
  panelsLog: PanelsLog,
  panels: PanelsMap,
  timestamp: number
) => {
  const newPanelsLog = { ...panelsLog };
  for (const [panelName, panel] of Object.entries(panels || {})) {
    const newPanelLog = [...(newPanelsLog[panelName] || [])];
    // If we have an existing panel log for the same status, don't log it
    if (
      newPanelLog.length > 0 &&
      newPanelLog[newPanelLog.length - 1].status === panel.status
    ) {
      continue;
    }
    newPanelLog.push(buildPanelLog(panel, timestamp));
    newPanelsLog[panelName] = newPanelLog;
  }

  return newPanelsLog;
};

const buildSelectedDashboardInputsFromSearchParams = (searchParams) => {
  const selectedDashboardInputs = {};
  // @ts-ignore
  for (const entry of searchParams.entries()) {
    if (!entry[0].startsWith("input")) {
      continue;
    }
    selectedDashboardInputs[entry[0]] = entry[1];
  }
  return selectedDashboardInputs;
};

const buildSqlDataMap = (panels: PanelsMap): SQLDataMap => {
  const sqlPaths = paths(panels, { leavesOnly: true }).filter((path) => {
    return path.toString().endsWith(".sql");
  });
  const sqlDataMap = {};
  for (const sqlPath of sqlPaths) {
    // @ts-ignore
    const sql: string = get(panels, sqlPath);
    const panelPath = sqlPath.toString().substring(0, sqlPath.indexOf(".sql"));
    const panel = get(panels, panelPath);
    const data = panel.data;
    if (!data) {
      continue;
    }
    const args = panel.args;
    const key = getSqlDataMapKey(sql, args);
    if (!sqlDataMap[key]) {
      sqlDataMap[key] = data;
    }
  }
  return sqlDataMap;
};

const getSqlDataMapKey = (sql: string, args?: any[]) => {
  return `sql:${sql}${
    args && args.length > 0
      ? `:args:${args.map((a) => a.toString()).join(",")}`
      : ""
  }`;
};

const updateSelectedDashboard = (
  selectedDashboard: AvailableDashboard | null,
  newDashboards: AvailableDashboard[]
) => {
  if (!selectedDashboard) {
    return null;
  }
  const matchingDashboard = newDashboards.find(
    (dashboard) => dashboard.full_name === selectedDashboard.full_name
  );
  if (matchingDashboard) {
    return matchingDashboard;
  } else {
    return null;
  }
};

const wrapDefinitionInArtificialDashboard = (
  definition: PanelDefinition,
  layout: any
): DashboardDefinition => {
  const { title: defTitle, ...definitionWithoutTitle } = definition;
  const { title: layoutTitle, ...layoutWithoutTitle } = layout;
  return {
    artificial: true,
    name: definition.name,
    title: definition.title,
    panel_type: "dashboard",
    children: [
      {
        ...definitionWithoutTitle,
        ...layoutWithoutTitle,
      },
    ],
    dashboard: definition.dashboard,
  };
};

export {
  addDataToPanels,
  addPanelLog,
  buildDashboards,
  buildPanelsLog,
  buildSelectedDashboardInputsFromSearchParams,
  buildSqlDataMap,
  updatePanelsLogFromCompletedPanels,
  updateSelectedDashboard,
  wrapDefinitionInArtificialDashboard,
};
