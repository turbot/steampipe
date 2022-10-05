import { BasePrimitiveProps, ExecutablePrimitiveProps } from "../common";
import { CategoryMap, NodeAndEdgeProperties } from "../common/types";
import { ComponentType } from "react";
import { NodeAndEdgeData } from "../graphs/types";

export type BaseHierarchyProps = BasePrimitiveProps & ExecutablePrimitiveProps;

export type HierarchyProperties = NodeAndEdgeProperties & {
  categories?: CategoryMap;
};

export interface HierarchyProps extends BaseHierarchyProps {
  data?: NodeAndEdgeData;
  display_type?: HierarchyType;
  properties?: HierarchyProperties;
}

export type HierarchyType = "table" | "tree";

export interface IHierarchy {
  type: HierarchyType;
  component: ComponentType<any>;
}
