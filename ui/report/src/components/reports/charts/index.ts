import BarChart from "./BarChart";
import ColumnChart from "./ColumnChart";
import DonutChart from "./DonutChart";
import LineChart from "./LineChart";
import PieChart from "./PieChart";
import React from "react";
import Table from "../Table";
import { BasePrimitiveProps, ExecutablePrimitiveProps } from "../common";

export type BaseChartProps = BasePrimitiveProps & ExecutablePrimitiveProps;

type ChartLabelOptions = {
  display: "auto" | "always" | "none";
  format: string;
};

type ChartLegendOptions = {
  display: "auto" | "always" | "none";
  position: "top" | "right" | "bottom" | "left";
};

type ChartSeriesOptions = {
  title: string;
  color: string;
};

type ChartSeries = {
  [series: string]: ChartSeriesOptions;
};

type ChartAxisTitleOptions = {
  display: "always" | "none";
  align: "start" | "center" | "end";
  value: string;
};

type ChartXAxisOptions = {
  display: "auto" | "always" | "none";
  title: ChartAxisTitleOptions;
  labels: ChartLabelOptions;
};

type ChartYAxisOptions = {
  display: "auto" | "always" | "none";
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
  type: ChartType;
  axes?: ChartAxes;
  legend?: ChartLegendOptions;
  series?: ChartSeries;
  grouping?: ChartGrouping;
};

export type ChartProps = BaseChartProps & {
  properties: ChartProperties;
};

export type ChartType = "bar" | "column" | "donut" | "line" | "pie" | "table";

export type ChartJSType = "bar" | "doughnut" | "line" | "pie";

export interface IChart {
  type: ChartType;
  component: React.ComponentType<any>;
}

const TableWrapper: IChart = {
  type: "table",
  component: Table,
};

const charts = {
  [BarChart.type]: BarChart,
  [ColumnChart.type]: ColumnChart,
  [DonutChart.type]: DonutChart,
  [LineChart.type]: LineChart,
  [PieChart.type]: PieChart,
  [TableWrapper.type]: TableWrapper,
};

export default charts;
