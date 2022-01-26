import Chart from "../Chart";
import { ChartProps, IChart } from "../index";

const PieChart = (props: ChartProps) => (
  <Chart data={props.data} inputs={props.properties} />
);

const definition: IChart = {
  type: "pie",
  component: PieChart,
};

export default definition;
