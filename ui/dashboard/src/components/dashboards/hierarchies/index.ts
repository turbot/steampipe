import React from "react";
import Table from "../Table";
import Tree from "./Tree";
import {
  BasePrimitiveProps,
  ColorOverride,
  ExecutablePrimitiveProps,
} from "../common";

export type BaseChartProps = BasePrimitiveProps & ExecutablePrimitiveProps;

interface HierarchyCategoryOptions {
  title?: string;
  color?: ColorOverride;
}

export type HierarchyCategories = {
  [category: string]: HierarchyCategoryOptions;
};

export type HierarchyProperties = {
  type: HierarchyType;
  categories?: HierarchyCategories;
};

export type HierarchyProps = BaseChartProps & {
  properties?: HierarchyProperties;
};

export type HierarchyType = "table" | "tree";

export interface IHierarchy {
  type: HierarchyType;
  component: React.ComponentType<any>;
}

const TableWrapper: IHierarchy = {
  type: "table",
  component: Table,
};

const hierarchies = {
  [TableWrapper.type]: TableWrapper,
  [Tree.type]: Tree,
};

export default hierarchies;
