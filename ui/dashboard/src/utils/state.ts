import dayjs from "dayjs";
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
  DependencyPanelProperties,
  PanelDefinition,
  PanelLog,
  PanelsLog,
  PanelsMap,
  SQLDataMap,
} from "../types";
import {
  EdgeProperties,
  KeyValueStringPairs,
  NodeProperties,
} from "../components/dashboards/common/types";

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

const panelLogTitle = (panel: PanelDefinition) => {
  switch (panel.panel_type) {
    case "with":
      const dependencyPanelProperties = (panel.properties ||
        {}) as DependencyPanelProperties;
      return panel.title
        ? panel.title
        : dependencyPanelProperties && dependencyPanelProperties.name
        ? dependencyPanelProperties.name
        : panel.name;
    case "edge":
      if (panel.title) {
        return panel.title;
      }
      const edgeProperties = (panel.properties || {}) as EdgeProperties;
      if (edgeProperties.category) {
        if (edgeProperties.category.title) {
          return edgeProperties.category.title as string;
        }
        return edgeProperties.category.name as string;
      }
      return panel.name;
    case "node":
      if (panel.title) {
        return panel.title;
      }
      const nodeProperties = (panel.properties || {}) as NodeProperties;
      if (nodeProperties.category) {
        if (nodeProperties.category.title) {
          return nodeProperties.category.title as string;
        }
        return nodeProperties.category.name as string;
      }
      return panel.name;
    default:
      return panel.title || panel.name;
  }
};

const buildPanelLog = (
  panel: PanelDefinition,
  timestamp: string,
  executionTime?: number
): PanelLog => {
  return {
    error: panel.status === "error" ? panel.error : null,
    executionTime,
    status: panel.status as DashboardRunState,
    timestamp,
    title: panelLogTitle(panel),
  };
};

const buildPanelsLog = (panels: PanelsMap, timestamp: string) => {
  const panelsLog: PanelsLog = {};
  for (const [name, panel] of Object.entries(panels || {})) {
    panelsLog[name] = [buildPanelLog(panel, timestamp)];
  }
  return panelsLog;
};

const calculateExecutionTime = (
  timestamp: string,
  panel: PanelDefinition,
  panelLogs: PanelLog[]
): number | undefined => {
  let overallTime: number | undefined = undefined;
  if (panel.status === "complete") {
    const runningLog = panelLogs.find((l) => l.status === "running");
    if (runningLog) {
      overallTime = dayjs(timestamp).diff(runningLog.timestamp);
    }
  }
  return overallTime;
};

const addUpdatedPanelLogs = (
  panelsLog: PanelsLog,
  panel: PanelDefinition,
  timestamp: string
) => {
  const newPanelsLog = { ...panelsLog };
  const newPanelLog = [...(newPanelsLog[panel.name] || [])];
  if (newPanelLog.find((l) => l.status === panel.status)) {
    return newPanelsLog;
  } else {
    const overallTime = calculateExecutionTime(timestamp, panel, newPanelLog);
    newPanelLog.push(buildPanelLog(panel, timestamp, overallTime));
  }
  newPanelsLog[panel.name] = newPanelLog;
  return newPanelsLog;
};

const updatePanelsLogFromCompletedPanels = (
  panelsLog: PanelsLog,
  panels: PanelsMap,
  timestamp: string
) => {
  const newPanelsLog = { ...panelsLog };
  for (const [panelName, panel] of Object.entries(panels || {})) {
    const newPanelLog = [...(newPanelsLog[panelName] || [])];
    // If we have an existing panel log for the same status, don't log it
    if (newPanelLog.find((l) => l.status === panel.status)) {
      continue;
    }
    const overallTime = calculateExecutionTime(timestamp, panel, newPanelLog);
    newPanelLog.push(buildPanelLog(panel, timestamp, overallTime));
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
  addUpdatedPanelLogs,
  buildDashboards,
  buildPanelsLog,
  buildSelectedDashboardInputsFromSearchParams,
  buildSqlDataMap,
  panelLogTitle,
  updatePanelsLogFromCompletedPanels,
  updateSelectedDashboard,
  wrapDefinitionInArtificialDashboard,
};
