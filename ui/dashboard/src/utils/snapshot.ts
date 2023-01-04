import {
  EXECUTION_SCHEMA_VERSION_20220614,
  EXECUTION_SCHEMA_VERSION_20220929,
  EXECUTION_SCHEMA_VERSION_20221222,
} from "../constants/versions";
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
    case EXECUTION_SCHEMA_VERSION_20220614:
    case EXECUTION_SCHEMA_VERSION_20220929:
    case EXECUTION_SCHEMA_VERSION_20221222:
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
