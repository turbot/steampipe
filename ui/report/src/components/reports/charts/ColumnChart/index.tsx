import Chart from "../Chart";
import React from "react";
import { ChartProps, IChart } from "../index";

const ColumnChart = (props: ChartProps) => (
  <Chart data={props.data} inputs={{ type: "column", ...props.properties }} />
);

const definition: IChart = {
  type: "column",
  component: ColumnChart,
};

export default definition;
