import Chart from "../Chart";
import { ChartProps, IChart } from "../index";

const LineChart = (props: ChartProps) => (
  <Chart data={props.data} inputs={props.properties} />
);

const definition: IChart = {
  type: "line",
  component: LineChart,
};

export default definition;
