import { BasePrimitiveProps, ExecutablePrimitiveProps } from "../common";
import { CategoryMap, NodeAndEdgeProperties } from "../common/types";
import { ComponentType } from "react";
import { NodeAndEdgeData } from "../graphs/types";
import { PanelDefinition } from "../../../types";

export type BaseFlowProps = PanelDefinition &
  BasePrimitiveProps &
  ExecutablePrimitiveProps;

export type FlowProperties = NodeAndEdgeProperties;

export type FlowProps = BaseFlowProps & {
  categories: CategoryMap;
  data?: NodeAndEdgeData;
  display_type?: FlowType;
  properties?: NodeAndEdgeProperties;
};

export type FlowType = "sankey" | "table";

export type IFlow = {
  type: FlowType;
  component: ComponentType<any>;
};
