import React from "react";
import Sankey from "./Sankey";
import Table from "../Table";
import { BasePrimitiveProps, ExecutablePrimitiveProps } from "../common";

export type BaseChartProps = BasePrimitiveProps & ExecutablePrimitiveProps;

export type HierarchyProperties = {
  type: HierarchyType;
};

export type HierarchyProps = BaseChartProps & {
  properties?: HierarchyProperties;
};

export type HierarchyType = "sankey" | "table";

export type EChartsType = "sankey";

export interface IHierarchy {
  type: HierarchyType;
  component: React.ComponentType<any>;
}

const TableWrapper: IHierarchy = {
  type: "table",
  component: Table,
};

const hierarchies = {
  [Sankey.type]: Sankey,
  [TableWrapper.type]: TableWrapper,
};

export default hierarchies;
