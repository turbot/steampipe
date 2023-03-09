import {
  BasePrimitiveProps,
  ExecutablePrimitiveProps,
  LeafNodeDataColumn,
} from "../common";
import {
  Category,
  CategoryMap,
  KeyValuePairs,
  NodeAndEdgeProperties,
} from "../common/types";
import { ComponentType } from "react";
import { DashboardRunState, PanelDefinition } from "../../../types";

export type NodeAndEdgeDataFormat = "LEGACY" | "NODE_AND_EDGE";

export type BaseGraphProps = PanelDefinition &
  BasePrimitiveProps &
  ExecutablePrimitiveProps;

export type NodeAndEdgeDataRow = {
  from_id?: string;
  to_id?: string;
  id?: string;
  title?: string;
  category?: string;
  properties?: KeyValuePairs;
  depth?: number;
};

export type NodeAndEdgeDataColumn = LeafNodeDataColumn;

export type NodeAndEdgeData = {
  columns: NodeAndEdgeDataColumn[];
  rows: NodeAndEdgeDataRow[];
};

export type DagreRankDir = "LR" | "TB";

export type GraphDirection = "left_right" | "top_down" | "LR" | "TB";

export type GraphProperties = NodeAndEdgeProperties & {
  direction?: GraphDirection | null;
};

export type GraphProps = BaseGraphProps & {
  categories: CategoryMap;
  data?: NodeAndEdgeData;
  display_type?: GraphType;
  properties?: GraphProperties;
};

export type GraphType = "graph" | "table";

export type IGraph = {
  type: GraphType;
  component: ComponentType<any>;
};

type BaseNodeAndEdgeStatus = {
  id: string;
  state: DashboardRunState;
  category?: Category;
  title?: string;
  error?: string;
};

export type WithStatusMap = {
  [name: string]: WithStatus;
};

export type NodeStatus = BaseNodeAndEdgeStatus & {
  dependencies?: string[];
};

export type EdgeStatus = BaseNodeAndEdgeStatus & {
  dependencies?: string[];
};

export type WithStatus = BaseNodeAndEdgeStatus & {
  title?: string;
};

export type NodeAndEdgeStatus = {
  withs: WithStatusMap;
  nodes: NodeStatus[];
  edges: EdgeStatus[];
};

export type GraphStatuses = {
  [key in DashboardRunState]: {
    total: number;
    withs: WithStatus[];
    nodes: NodeStatus[];
    edges: EdgeStatus[];
  };
};
