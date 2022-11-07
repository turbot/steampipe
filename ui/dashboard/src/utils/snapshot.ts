import { PanelDefinition } from "../types";

const stripObjectProperties = (obj) => {
  if (!obj) {
    return {};
  }
  const {
    documentation,
    search_path,
    search_path_prefix,
    source_definition,
    sql,
    ...rest
  } = obj;

  return { ...rest };
};

const stripSnapshotDataForExport = (snapshot) => {
  if (!snapshot) {
    return {};
  }

  switch (snapshot.schema_version) {
    case "20220614":
    case "20220929":
      const { panels, ...restSnapshot } = stripObjectProperties(snapshot);
      const newPanels = {};
      for (const [name, panel] of Object.entries(panels)) {
        const { properties, ...restPanel } = stripObjectProperties(
          panel
        ) as PanelDefinition;
        const newPanel: PanelDefinition = {
          ...restPanel,
        };
        if (properties) {
          newPanel.properties = stripObjectProperties(properties);
        }
        newPanels[name] = newPanel;
      }

      return {
        ...restSnapshot,
        panels: newPanels,
      };
    default:
      throw new Error(
        `Unsupported dashboard event schema ${snapshot.schema_version}`
      );
  }
};

export { stripSnapshotDataForExport };
