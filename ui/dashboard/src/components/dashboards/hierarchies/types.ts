import { BaseCategoryOptions } from "../common/types";
import {
  BasePrimitiveProps,
  ColorOverride,
  ExecutablePrimitiveProps,
} from "../common";
import { ComponentType } from "react";

export type BaseChartProps = BasePrimitiveProps & ExecutablePrimitiveProps;

interface HierarchyCategoryOptions extends BaseCategoryOptions {
  title?: string;
  color?: ColorOverride;
}

export type HierarchyCategories = {
  [category: string]: HierarchyCategoryOptions;
};

export type HierarchyProperties = {
  categories?: HierarchyCategories;
};

export type HierarchyProps = BaseChartProps & {
  display_type?: HierarchyType;
  properties?: HierarchyProperties;
};

export type HierarchyType = "table" | "tree";

export interface IHierarchy {
  type: HierarchyType;
  component: ComponentType<any>;
}
