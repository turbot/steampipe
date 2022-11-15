import { BasePrimitiveProps, ExecutablePrimitiveProps } from "../common";
import { ComponentType } from "react";
import { NodeAndEdgeData } from "../graphs/types";
import { CategoryMap, NodeAndEdgeProperties } from "../common/types";

export type BaseFlowProps = BasePrimitiveProps & ExecutablePrimitiveProps;

export type FlowProperties = NodeAndEdgeProperties;

export interface FlowProps extends BaseFlowProps {
  categories: CategoryMap;
  data?: NodeAndEdgeData;
  display_type?: FlowType;
  properties?: NodeAndEdgeProperties;
}

export type FlowType = "sankey" | "table";

export interface IFlow {
  type: FlowType;
  component: ComponentType<any>;
}
