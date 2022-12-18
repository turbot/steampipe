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

export type NodeAndEdgeDataFormat = "LEGACY" | "NODE_AND_EDGE";

export type BaseGraphProps = BasePrimitiveProps & ExecutablePrimitiveProps;

export interface NodeAndEdgeDataRow {
  from_id?: string;
  to_id?: string;
  id?: string;
  title?: string;
  category?: string;
  properties?: KeyValuePairs;
  depth?: number;
}

export type NodeAndEdgeDataColumn = LeafNodeDataColumn;

export interface NodeAndEdgeData {
  columns: NodeAndEdgeDataColumn[];
  rows: NodeAndEdgeDataRow[];
}

export type DagreRankDir = "LR" | "TB";

export type GraphDirection = "left_right" | "top_down" | "LR" | "TB";

export type GraphProperties = NodeAndEdgeProperties & {
  direction?: GraphDirection | null;
};

export interface GraphProps extends BaseGraphProps {
  categories: CategoryMap;
  data?: NodeAndEdgeData;
  display_type?: GraphType;
  properties?: GraphProperties;
}

export type GraphType = "graph" | "table";

export interface IGraph {
  type: GraphType;
  component: ComponentType<any>;
}

export type NodeAndEdgeState = "pending" | "error" | "complete";

type BaseNodeAndEdgeStatus = {
  id: string;
  state: NodeAndEdgeState;
  category?: Category;
  title?: string;
  error?: string;
};

export type WithStatusMap = {
  [name: string]: WithStatus;
};

export type NodeStatus = BaseNodeAndEdgeStatus & {
  withs?: string[];
};

export type EdgeStatus = BaseNodeAndEdgeStatus & {
  withs?: string[];
};

export type WithStatus = BaseNodeAndEdgeStatus & {
  title?: string;
};

export type NodeAndEdgeStatus = {
  withs: WithStatusMap;
  nodes: NodeStatus[];
  edges: EdgeStatus[];
};
