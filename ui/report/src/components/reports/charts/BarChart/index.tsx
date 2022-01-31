import Chart from "../Chart";
import React from "react";
import { ChartProps, IChart } from "../index";

const BarChart = (props: ChartProps) => (
  <Chart data={props.data} inputs={{ type: "bar", ...props.properties }} />
);

const definition: IChart = {
  type: "bar",
  component: BarChart,
};

export default definition;
