import { BasePrimitiveProps, ExecutablePrimitiveProps } from "../common";
import { ComponentType } from "react";
import { CategoryMap, NodeAndEdgeProperties } from "../common/types";
import { NodeAndEdgeData } from "../graphs/types";

export type BaseHierarchyProps = BasePrimitiveProps & ExecutablePrimitiveProps;

export type HierarchyProperties = NodeAndEdgeProperties;

export interface HierarchyProps extends BaseHierarchyProps {
  categories: CategoryMap;
  data?: NodeAndEdgeData;
  display_type?: HierarchyType;
  properties?: NodeAndEdgeProperties;
}

export type HierarchyType = "table" | "tree";

export interface IHierarchy {
  type: HierarchyType;
  component: ComponentType<any>;
}
