import usePanelControls from "./usePanelControls";
import { BaseChartProps } from "../components/dashboards/charts/types";
import { CardProps } from "../components/dashboards/Card";
import { createContext, ReactNode, useContext, useMemo, useState } from "react";
import { FlowProps } from "../components/dashboards/flows/types";
import { GraphProps } from "../components/dashboards/graphs/types";
import { HierarchyProps } from "../components/dashboards/hierarchies/types";
import { ImageProps } from "../components/dashboards/Image";
import { InputProps } from "../components/dashboards/inputs/types";
import { IPanelControl } from "../components/dashboards/layout/Panel/PanelControls";
import { PanelDefinition, PanelDependenciesByStatus } from "../types";
import { TableProps } from "../components/dashboards/Table";
import { TextProps } from "../components/dashboards/Text";
import { useDashboard } from "./useDashboard";

type IPanelContext = {
  definition:
    | BaseChartProps
    | CardProps
    | FlowProps
    | GraphProps
    | HierarchyProps
    | ImageProps
    | InputProps
    | PanelDefinition
    | TableProps
    | TextProps;
  dependencies: PanelDefinition[];
  dependenciesByStatus: PanelDependenciesByStatus;
  panelControls: IPanelControl[];
  panelInformation: ReactNode | null;
  showPanelControls: boolean;
  showPanelInformation: boolean;
  setPanelInformation: (information: ReactNode) => void;
  setShowPanelControls: (show: boolean) => void;
  setShowPanelInformation: (show: boolean) => void;
};

type PanelProviderProps = {
  children: ReactNode;
  definition:
    | BaseChartProps
    | CardProps
    | FlowProps
    | GraphProps
    | HierarchyProps
    | ImageProps
    | InputProps
    | PanelDefinition
    | TableProps
    | TextProps;
  showControls?: boolean;
};

const PanelContext = createContext<IPanelContext | null>(null);

const PanelProvider = ({
  children,
  definition,
  showControls,
}: PanelProviderProps) => {
  const { panelsMap } = useDashboard();
  const [showPanelControls, setShowPanelControls] = useState(false);
  const [showPanelInformation, setShowPanelInformation] = useState(false);
  const [panelInformation, setPanelInformation] = useState<ReactNode | null>(
    null
  );
  const { panelControls } = usePanelControls(definition, showControls);
  const dependencies = useMemo<PanelDefinition[]>(() => {
    if (
      !definition ||
      !definition.dependencies ||
      definition.dependencies.length === 0
    ) {
      return [];
    }
    const dependencies: PanelDefinition[] = [];
    for (const dependency of definition.dependencies) {
      const dependencyPanel = panelsMap[dependency];
      if (!dependencyPanel) {
        continue;
      }
      dependencies.push(dependencyPanel);
    }
    return dependencies;
  }, [definition, panelsMap]);

  const dependenciesByStatus = useMemo<PanelDependenciesByStatus>(() => {
    if (
      !definition ||
      !definition.dependencies ||
      definition.dependencies.length === 0
    ) {
      return {} as PanelDependenciesByStatus;
    }
    const dependenciesByStatus: PanelDependenciesByStatus = {};
    for (const dependency of definition.dependencies) {
      const dependencyPanel = panelsMap[dependency];
      if (!dependencyPanel || !dependencyPanel.status) {
        continue;
      }
      const statuses = dependenciesByStatus[dependencyPanel.status] || [];
      statuses.push(dependencyPanel);
      dependenciesByStatus[dependencyPanel.status] = statuses;
    }
    return dependenciesByStatus;
  }, [definition, panelsMap]);

  return (
    <PanelContext.Provider
      value={{
        definition,
        dependencies,
        dependenciesByStatus,
        panelControls,
        panelInformation,
        showPanelControls,
        showPanelInformation,
        setPanelInformation,
        setShowPanelControls,
        setShowPanelInformation,
      }}
    >
      {children}
    </PanelContext.Provider>
  );
};

const usePanel = () => {
  const context = useContext(PanelContext);
  if (context === undefined) {
    throw new Error("usePanel must be used within a PanelContext");
  }
  return context as IPanelContext;
};

export { PanelContext, PanelProvider, usePanel };
