import {
  BasePrimitiveProps,
  ColorOverride,
  ExecutablePrimitiveProps,
} from "../common";
import { ComponentType } from "react";
import { PanelDefinition } from "../../../types";

export type BaseChartProps = PanelDefinition &
  BasePrimitiveProps &
  ExecutablePrimitiveProps;

type ChartLabelOptions = {
  display: "auto" | "all" | "none";
  format: string;
};

type ChartLegendOptions = {
  display: "auto" | "all" | "none";
  position: "top" | "right" | "bottom" | "left";
};

type ChartSeriesPointOptions = {
  name: string;
  color?: ColorOverride;
};

type ChartSeriesPoints = {
  [key: string]: ChartSeriesPointOptions;
};

export type ChartSeriesOptions = {
  title: string;
  color?: ColorOverride;
  points: ChartSeriesPoints;
};

export type ChartSeries = {
  [series: string]: ChartSeriesOptions;
};

export type ChartTransform = "auto" | "crosstab" | "none";

type ChartAxisTitleOptions = {
  display: "all" | "none";
  align: "start" | "center" | "end";
  value: string;
};

type ChartXAxisOptions = {
  display: "auto" | "all" | "none";
  title: ChartAxisTitleOptions;
  labels: ChartLabelOptions;
  min: number;
  max: number;
};

type ChartYAxisOptions = {
  display: "auto" | "all" | "none";
  title: ChartAxisTitleOptions;
  labels: ChartLabelOptions;
  min: number;
  max: number;
};

type ChartAxes = {
  x: ChartXAxisOptions;
  y: ChartYAxisOptions;
};

type ChartGrouping = "stack" | "compare";

export type ChartProperties = {
  axes?: ChartAxes;
  legend?: ChartLegendOptions;
  series?: ChartSeries;
  transform?: ChartTransform;
  grouping?: ChartGrouping;
};

export type ChartProps = BaseChartProps & {
  display_type?: ChartType;
  properties?: ChartProperties;
};

export type ChartType =
  | "area"
  | "bar"
  | "column"
  | "donut"
  | "line"
  | "pie"
  | "table";

export type IChart = {
  type: ChartType;
  component: ComponentType<any>;
};
