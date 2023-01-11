import usePanelControls from "./usePanelControls";
import { BaseChartProps } from "../components/dashboards/charts/types";
import { CardProps } from "../components/dashboards/Card";
import { createContext, ReactNode, useContext, useMemo, useState } from "react";
import {
  DashboardInputs,
  DashboardRunState,
  PanelDefinition,
  PanelDependenciesByStatus,
  PanelsMap,
} from "../types";
import { FlowProps } from "../components/dashboards/flows/types";
import { getNodeAndEdgeDataFormat } from "../components/dashboards/common/useNodeAndEdgeData";
import { GraphProps } from "../components/dashboards/graphs/types";
import { HierarchyProps } from "../components/dashboards/hierarchies/types";
import { ImageProps } from "../components/dashboards/Image";
import {
  InputProperties,
  InputProps,
} from "../components/dashboards/inputs/types";
import { IPanelControl } from "../components/dashboards/layout/Panel/PanelControls";
import { NodeAndEdgeProperties } from "../components/dashboards/common/types";
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
  inputPanelsAwaitingValue: PanelDefinition[];
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

const recordDependency = (
  definition: PanelDefinition,
  panelsMap: PanelsMap,
  selectedDashboardInputs: DashboardInputs,
  dependencies: PanelDefinition[],
  dependenciesByStatus: PanelDependenciesByStatus,
  inputPanelsAwaitingValue: PanelDefinition[],
  recordedInputPanels: {}
) => {
  // Record this panel as a dependency
  dependencies.push(definition);

  // Keep track of this panel by its status
  const statuses =
    dependenciesByStatus[definition.status as DashboardRunState] || [];
  statuses.push(definition);
  dependenciesByStatus[definition.status as DashboardRunState] = statuses;

  // Is this panel an input? If so, does it have a value?
  const isInput = definition.panel_type === "input";
  const inputProperties = isInput
    ? (definition.properties as InputProperties)
    : null;
  const hasInputValue =
    isInput &&
    inputProperties?.unqualified_name &&
    !!selectedDashboardInputs[inputProperties?.unqualified_name];
  if (isInput && !hasInputValue && !recordedInputPanels[definition.name]) {
    inputPanelsAwaitingValue.push(definition);
    recordedInputPanels[definition.name] = definition;
  }

  for (const dependency of definition?.dependencies || []) {
    const dependencyPanel = panelsMap[dependency];
    if (!dependencyPanel || !dependencyPanel.status) {
      continue;
    }
    recordDependency(
      dependencyPanel,
      panelsMap,
      selectedDashboardInputs,
      dependencies,
      dependenciesByStatus,
      inputPanelsAwaitingValue,
      recordedInputPanels
    );
  }
};

const PanelProvider = ({
  children,
  definition,
  showControls,
}: PanelProviderProps) => {
  const { selectedDashboardInputs, panelsMap } = useDashboard();
  const [showPanelControls, setShowPanelControls] = useState(false);
  const [showPanelInformation, setShowPanelInformation] = useState(false);
  const [panelInformation, setPanelInformation] = useState<ReactNode | null>(
    null
  );
  const { panelControls } = usePanelControls(definition, showControls);
  const { dependencies, dependenciesByStatus, inputPanelsAwaitingValue } =
    useMemo(() => {
      if (!definition) {
        return {
          dependencies: [],
          dependenciesByStatus: {},
          inputPanelsAwaitingValue: [],
        };
      }
      const dataFormat = getNodeAndEdgeDataFormat(
        definition.properties as NodeAndEdgeProperties
      );
      if (
        dataFormat === "LEGACY" &&
        (!definition.dependencies || definition.dependencies.length === 0)
      ) {
        return {
          dependencies: [],
          dependenciesByStatus: {},
          inputPanelsAwaitingValue: [],
        };
      }
      const dependencies: PanelDefinition[] = [];
      const dependenciesByStatus: PanelDependenciesByStatus = {};
      const inputPanelsAwaitingValue: PanelDefinition[] = [];
      const recordedInputPanels = {};

      if (dataFormat === "NODE_AND_EDGE") {
        const nodeAndEdgeProperties =
          definition.properties as NodeAndEdgeProperties;
        for (const node of nodeAndEdgeProperties.nodes || []) {
          const nodePanel = panelsMap[node];
          if (!nodePanel || !nodePanel.status) {
            continue;
          }
          recordDependency(
            nodePanel,
            panelsMap,
            selectedDashboardInputs,
            dependencies,
            dependenciesByStatus,
            inputPanelsAwaitingValue,
            recordedInputPanels
          );
        }
        for (const edge of nodeAndEdgeProperties.edges || []) {
          const edgePanel = panelsMap[edge];
          if (!edgePanel || !edgePanel.status) {
            continue;
          }
          recordDependency(
            edgePanel,
            panelsMap,
            selectedDashboardInputs,
            dependencies,
            dependenciesByStatus,
            inputPanelsAwaitingValue,
            recordedInputPanels
          );
        }
      }

      for (const dependency of definition.dependencies || []) {
        const dependencyPanel = panelsMap[dependency];
        if (!dependencyPanel || !dependencyPanel.status) {
          continue;
        }
        recordDependency(
          dependencyPanel,
          panelsMap,
          selectedDashboardInputs,
          dependencies,
          dependenciesByStatus,
          inputPanelsAwaitingValue,
          recordedInputPanels
        );
      }

      if (
        definition.name ===
          "aws_insights.graph.container_dashboard_s3_bucket_detail_anonymous_container_1_anonymous_graph_0" &&
        dataFormat === "NODE_AND_EDGE"
      ) {
        console.log({
          status: definition.status,
          dependencies,
          dependenciesByStatus,
          inputPanelsAwaitingValue,
        });
      }

      return { dependencies, dependenciesByStatus, inputPanelsAwaitingValue };
    }, [definition, panelsMap, selectedDashboardInputs]);

  return (
    <PanelContext.Provider
      value={{
        definition,
        dependencies,
        dependenciesByStatus,
        inputPanelsAwaitingValue,
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
