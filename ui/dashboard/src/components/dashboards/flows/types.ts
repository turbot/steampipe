import { BasePrimitiveProps, ExecutablePrimitiveProps } from "../common";
import { CategoryMap, NodeAndEdgeProperties } from "../common/types";
import { ComponentType } from "react";
import { NodeAndEdgeData } from "../graphs/types";

export type BaseFlowProps = BasePrimitiveProps & ExecutablePrimitiveProps;

export type FlowProperties = NodeAndEdgeProperties & {
  categories?: CategoryMap;
};

export interface FlowProps extends BaseFlowProps {
  data?: NodeAndEdgeData;
  display_type?: FlowType;
  properties?: FlowProperties;
}

export type FlowType = "sankey" | "table";

export interface IFlow {
  type: FlowType;
  component: ComponentType<any>;
}
