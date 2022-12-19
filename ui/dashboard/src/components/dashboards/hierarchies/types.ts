import { BasePrimitiveProps, ExecutablePrimitiveProps } from "../common";
import { ComponentType } from "react";
import { CategoryMap, NodeAndEdgeProperties } from "../common/types";
import { NodeAndEdgeData } from "../graphs/types";
import { PanelDefinition } from "../../../types";

export type BaseHierarchyProps = PanelDefinition &
  BasePrimitiveProps &
  ExecutablePrimitiveProps;

export type HierarchyProperties = NodeAndEdgeProperties;

export type HierarchyProps = BaseHierarchyProps & {
  categories: CategoryMap;
  data?: NodeAndEdgeData;
  display_type?: HierarchyType;
  properties?: NodeAndEdgeProperties;
};

export type HierarchyType = "table" | "tree";

export type IHierarchy = {
  type: HierarchyType;
  component: ComponentType<any>;
};
