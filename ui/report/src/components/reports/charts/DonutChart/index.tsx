import Chart from "../Chart";
import { ChartProps, IChart } from "../index";

const DonutChart = (props: ChartProps) => (
  <Chart data={props.data} inputs={{ type: "donut", ...props.properties }} />
);

const definition: IChart = {
  type: "donut",
  component: DonutChart,
};

export default definition;
