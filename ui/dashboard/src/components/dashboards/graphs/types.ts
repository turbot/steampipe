import { BaseCategoryOptions } from "../common/types";
import {
  BasePrimitiveProps,
  ColorOverride,
  ExecutablePrimitiveProps,
} from "../common";
import { ComponentType } from "react";

export type BaseChartProps = BasePrimitiveProps & ExecutablePrimitiveProps;

interface GraphCategoryOptions extends BaseCategoryOptions {
  title?: string;
  color?: ColorOverride;
  depth?: number;
  icon?: string;
}

export type GraphCategories = {
  [category: string]: GraphCategoryOptions;
};

export type GraphProperties = {
  categories?: GraphCategories;
  direction: "LR" | "RL" | "TB" | "BT";
};

export type GraphProps = BaseChartProps & {
  display_type?: GraphType;
  properties?: GraphProperties;
};

export type GraphType = "graph" | "table";

export interface IGraph {
  type: GraphType;
  component: ComponentType<any>;
}
