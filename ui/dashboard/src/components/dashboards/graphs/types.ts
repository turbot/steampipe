import {
  BasePrimitiveProps,
  ExecutablePrimitiveProps,
  LeafNodeDataColumn,
} from "../common";
import { ComponentType } from "react";
import { KeyValuePairs, NodeAndEdgeProperties } from "../common/types";

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

export type GraphProperties = NodeAndEdgeProperties & {
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

export type NodeAndEdgeState = "pending" | "error" | "complete";

type BaseNodeAndEdgeStatus = {
  id: string;
  title: string;
  state: NodeAndEdgeState;
};

export type NodeStatus = BaseNodeAndEdgeStatus & {
  count: number;
};

export type EdgeStatus = BaseNodeAndEdgeStatus;

export type NodeAndEdgeStatus = {
  nodes: NodeStatus[];
  edges: EdgeStatus[];
};
