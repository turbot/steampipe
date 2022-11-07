import { PanelDefinition } from "../types";

const stripSnapshotDataForExport = (snapshot) => {
  if (!snapshot) {
    return {};
  }

  switch (snapshot.schema_version) {
    case "20220614":
    case "20220929":
      const { panels, search_path, search_path_prefix, ...rest } = snapshot;
      const newPanels = {};
      for (const [name, panel] of Object.entries(panels)) {
        const { documentation, sql, source_definition, ...rest } =
          panel as PanelDefinition;
        newPanels[name] = {
          ...rest,
        };
      }

      return {
        ...rest,
        panels: newPanels,
      };
    default:
      throw new Error(
        `Unsupported dashboard event schema ${snapshot.schema_version}`
      );
  }
};

export { stripSnapshotDataForExport };
