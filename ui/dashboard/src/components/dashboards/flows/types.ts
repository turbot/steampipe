import { BaseCategoryOptions } from "../common/types";
import {
  BasePrimitiveProps,
  ColorOverride,
  ExecutablePrimitiveProps,
} from "../common";
import { ComponentType } from "react";

export type BaseChartProps = BasePrimitiveProps & ExecutablePrimitiveProps;

interface FlowCategoryOptions extends BaseCategoryOptions {
  title?: string;
  color?: ColorOverride;
  depth?: number;
  icon?: string;
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
  component: ComponentType<any>;
}
