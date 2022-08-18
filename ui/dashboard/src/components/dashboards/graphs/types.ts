import React from "react";
import {
  BasePrimitiveProps,
  ColorOverride,
  ExecutablePrimitiveProps,
} from "../common";

export type BaseChartProps = BasePrimitiveProps & ExecutablePrimitiveProps;

interface GraphCategoryFoldOption {
  threshold: number;
  title?: string;
  icon?: string;
}

interface GraphCategoryOptions {
  title?: string;
  color?: ColorOverride;
  depth?: number;
  icon?: string;
  fold?: GraphCategoryFoldOption;
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
  component: React.ComponentType<any>;
}
