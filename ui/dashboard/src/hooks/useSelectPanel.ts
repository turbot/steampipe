import { DashboardActions, PanelDefinition } from "../types";
import { useCallback } from "react";
import { useDashboard } from "./useDashboard";

const useSelectPanel = (definition: PanelDefinition) => {
  const { dispatch } = useDashboard();
  const openPanelDetail = useCallback(
    async (e) => {
      e.stopPropagation();
      dispatch({
        type: DashboardActions.SELECT_PANEL,
        panel: definition,
      });
    },
    [dispatch, definition]
  );

  return { select: openPanelDetail };
};

export default useSelectPanel;
