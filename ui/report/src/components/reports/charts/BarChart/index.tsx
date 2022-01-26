import Chart from "../Chart";
import React from "react";
import { ChartProps, IChart } from "../index";

const BarChart = (props: ChartProps) => (
  <Chart data={props.data} inputs={props.properties} />
);

const definition: IChart = {
  type: "bar",
  component: BarChart,
};

export default definition;
