import { TableColumnDisplay, TableColumnWrap } from "../Table";
import { ColorOverride, LeafNodeDataRow } from "./index";
import { Graph } from "graphlib";

export interface CategoryFields {
  [name: string]: CategoryField;
}

export interface CategoryField {
  name: string;
  href?: string | null;
  display?: TableColumnDisplay;
  wrap?: TableColumnWrap;
}

export interface KeyValuePairs {
  [key: string]: any;
}

export interface KeyValueStringPairs {
  [key: string]: string;
}

export interface NodeProperties {
  category?: Category;
}

export interface EdgeProperties {
  category?: Category;
}

export interface NodeAndEdgeProperties {
  categories?: CategoryMap;
  edges?: string[];
  nodes?: string[];
}

export interface CategoryFold {
  threshold: number;
  title?: string;
  icon?: string;
}

export interface Category {
  name?: string;
  color?: ColorOverride;
  depth?: number;
  fields?: CategoryFields;
  fold?: CategoryFold;
  href?: string;
  icon?: string;
  title?: string;
}

export interface CategoryMap {
  [category: string]: Category;
}

export interface FoldedNode {
  id: string;
  title?: string;
}

export interface Node {
  id: string;
  title: string | null;
  category: string | null;
  depth: number | null;
  row_data: LeafNodeDataRow | null;
  symbol: string | null;
  href: string | null;
  isFolded: boolean;
  foldedNodes?: FoldedNode[];
}

export interface Edge {
  id: string;
  from_id: string;
  to_id: string;
  title: string | null;
  category: string | null;
  row_data: LeafNodeDataRow | null;
  isFolded: boolean;
}

export interface NodeMap {
  [id: string]: Node;
}

export interface EdgeMap {
  [edge_id: string]: boolean;
}

export interface NodeCategoryMap {
  [category: string]: NodeMap;
}

interface NodesAndEdgesMetadata {
  has_multiple_roots: boolean;
  contains_duplicate_edges: boolean;
}

export interface NodesAndEdges {
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
}

export type TemplatesMap = {
  [key: string]: string;
};

export type RowRenderResult = {
  [key: string]: {
    result?: string;
    error?: string;
  };
};
