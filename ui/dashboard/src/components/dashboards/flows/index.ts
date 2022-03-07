import React from "react";
import Sankey from "./Sankey";
import Table from "../Table";
import {
  BasePrimitiveProps,
  ColorOverride,
  ExecutablePrimitiveProps,
} from "../common";

export type BaseChartProps = BasePrimitiveProps & ExecutablePrimitiveProps;

interface FlowCategoryOptions {
  title?: string;
  color?: ColorOverride;
}

export type FlowCategories = {
  [category: string]: FlowCategoryOptions;
};

export type FlowProperties = {
  type: FlowType;
  categories?: FlowCategories;
};

export type FlowProps = BaseChartProps & {
  properties?: FlowProperties;
};

export type FlowType = "sankey" | "table";

export interface IFlow {
  type: FlowType;
  component: React.ComponentType<any>;
}

const TableWrapper: IFlow = {
  type: "table",
  component: Table,
};

const flows = {
  [Sankey.type]: Sankey,
  [TableWrapper.type]: TableWrapper,
};

export default flows;
