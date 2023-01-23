import { ColorOverride, LeafNodeDataRow } from "./index";
import { DashboardRunState, PanelDefinition } from "../../../types";
import { Graph } from "graphlib";
import { TableColumnDisplay, TableColumnWrap } from "../Table";

export type CategoryProperties = {
  [name: string]: CategoryProperty;
};

export type CategoryProperty = {
  name: string;
  href?: string | null;
  display?: TableColumnDisplay;
  wrap?: TableColumnWrap;
};

export type KeyValuePairs = {
  [key: string]: any;
};

export type KeyValueStringPairs = {
  [key: string]: string;
};

export type NodeProperties = {
  name: string;
  category?: Category;
};

export type EdgeProperties = {
  name: string;
  category?: Category;
};

export type NodeAndEdgeProperties = {
  categories?: CategoryMap;
  edges?: string[];
  nodes?: string[];
};

export type CategoryFold = {
  threshold: number;
  title?: string;
  icon?: string;
};

export type Category = {
  name?: string;
  color?: ColorOverride;
  depth?: number;
  properties?: CategoryProperties;
  fold?: CategoryFold;
  href?: string;
  icon?: string;
  title?: string;
};

export type CategoryMap = {
  [category: string]: Category;
};

export type FoldedNode = {
  id: string;
  title?: string;
};

export type Node = {
  id: string;
  title: string | null;
  category: string | null;
  depth: number | null;
  row_data: LeafNodeDataRow | null;
  symbol: string | null;
  href: string | null;
  isFolded: boolean;
  foldedNodes?: FoldedNode[];
};

export type Edge = {
  id: string;
  from_id: string;
  to_id: string;
  title: string | null;
  category: string | null;
  row_data: LeafNodeDataRow | null;
  isFolded: boolean;
};

export type NodeMap = {
  [id: string]: Node;
};

export type EdgeMap = {
  [edge_id: string]: boolean;
};

export type NodeCategoryMap = {
  [category: string]: NodeMap;
};

type NodesAndEdgesMetadata = {
  has_multiple_roots: boolean;
  contains_duplicate_edges: boolean;
};

export type NodesAndEdges = {
  graph: Graph;
  nodes: Node[];
  edges: Edge[];
  nodeCategoryMap: NodeCategoryMap;
  nodeMap: KeyValuePairs;
  edgeMap: KeyValuePairs;
  root_nodes: NodeMap;
  categories: CategoryMap;
  metadata?: NodesAndEdgesMetadata;
  next_color_index?: number;
};

export type TemplatesMap = {
  [key: string]: string;
};

export type RowRenderResult = {
  [key: string]: {
    result?: string;
    error?: string;
  };
};

export type PanelDependencyByStatus = {
  [key in DashboardRunState]: {
    total: number;
    panels: PanelDefinition[];
  };
};

export type PanelDependencyStatuses = {
  status: PanelDependencyByStatus;
  inputsAwaitingValue: PanelDefinition[];
  total: number;
};
