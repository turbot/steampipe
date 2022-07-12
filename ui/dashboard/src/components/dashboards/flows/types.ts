import React from "react";
import {
  BasePrimitiveProps,
  ColorOverride,
  ExecutablePrimitiveProps,
} from "../common";

export type BaseChartProps = BasePrimitiveProps & ExecutablePrimitiveProps;

interface FlowCategoryOptions {
  title?: string;
  color?: ColorOverride;
  depth?: number;
}

export type FlowCategories = {
  [category: string]: FlowCategoryOptions;
};

export type FlowProperties = {
  categories?: FlowCategories;
};

export type FlowProps = BaseChartProps & {
  display_type?: FlowType;
  properties?: FlowProperties;
};

export type FlowType = "sankey" | "table";

export interface IFlow {
  type: FlowType;
  component: React.ComponentType<any>;
}
