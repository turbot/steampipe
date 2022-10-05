import {
  BasePrimitiveProps,
  ExecutablePrimitiveProps,
  LeafNodeDataColumn,
} from "../common";
import {
  CategoryMap,
  KeyValuePairs,
  NodeAndEdgeProperties,
} from "../common/types";
import { ComponentType } from "react";

export type BaseGraphProps = BasePrimitiveProps & ExecutablePrimitiveProps;

export interface NodeAndEdgeDataRow {
  from_id?: string;
  to_id?: string;
  id?: string;
  title?: string;
  category?: string;
  properties?: KeyValuePairs;
}

export type NodeAndEdgeDataColumn = LeafNodeDataColumn;

export interface NodeAndEdgeData {
  columns: NodeAndEdgeDataColumn[];
  rows: NodeAndEdgeDataRow[];
}

export type GraphProperties = NodeAndEdgeProperties & {
  categories?: CategoryMap;
  direction: "LR" | "RL" | "TB" | "BT";
};

export interface GraphProps extends BaseGraphProps {
  data?: NodeAndEdgeData;
  display_type?: GraphType;
  properties?: GraphProperties;
}

export type GraphType = "graph" | "table";

export interface IGraph {
  type: GraphType;
  component: ComponentType<any>;
}
