import get from "lodash/get";
import has from "lodash/has";
import paths from "deepdash/paths";
import set from "lodash/set";
import { DashboardDefinition, PanelsMap, SQLDataMap } from "../types/dashboard";
import { PanelDefinition } from "../types/panel";

const addDataToPanels = (
  panels: PanelsMap,
  sqlDataMap: SQLDataMap
): PanelsMap => {
  const sqlPaths = paths(panels, { leavesOnly: true }).filter((path) =>
    path.endsWith(".sql")
  );
  for (const sqlPath of sqlPaths) {
    // @ts-ignore
    const sql: string = get(panels, sqlPath);
    const data = sqlDataMap[sql];
    if (!data) {
      continue;
    }
    const dataPath = `${sqlPath.substring(0, sqlPath.indexOf(".sql"))}.data`;
    set(panels, dataPath, data);
  }
  return panels;
};

const buildSqlDataMap = (panels: PanelsMap): SQLDataMap => {
  const sqlPaths = paths(panels, { leavesOnly: true }).filter((path) =>
    path.endsWith(".sql")
  );
  const sqlDataMap = {};
  for (const sqlPath of sqlPaths) {
    // @ts-ignore
    const sql: string = get(panels, sqlPath);
    const dataPath = `${sqlPath.substring(0, sqlPath.indexOf(".sql"))}.data`;
    const data = get(panels, dataPath);
    if (!sqlDataMap[sql]) {
      sqlDataMap[sql] = data;
    }
  }
  return sqlDataMap;
};

const calculateProgress = (panelsMap) => {
  const panels: PanelDefinition[] = Object.values(panelsMap || {});
  let dataPanels = 0;
  let completeDataPanels = 0;
  for (const panel of panels) {
    const isControl = panel.panel_type === "control";
    const isDataPanel = has(panel, "sql");
    if (isControl || isDataPanel) {
      dataPanels += 1;
    }
    if (
      (isControl &&
        (panel.status === "complete" || panel.status === "error")) ||
      (isDataPanel && has(panel, "data"))
    ) {
      completeDataPanels += 1;
    }
  }
  if (dataPanels === 0) {
    return 100;
  }
  return Math.min(Math.ceil((completeDataPanels / dataPanels) * 100), 100);
};

const updatePanelsMapWithControlEvent = (panelsMap, action) => {
  return {
    ...panelsMap,
    [action.control.name]: action.control,
  };
};

const wrapDefinitionInArtificialDashboard = (
  definition: DashboardDefinition,
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
  buildSqlDataMap,
  calculateProgress,
  updatePanelsMapWithControlEvent,
  wrapDefinitionInArtificialDashboard,
};
