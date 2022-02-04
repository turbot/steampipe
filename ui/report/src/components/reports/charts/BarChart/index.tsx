import Chart from "../Chart";
import { ChartProps, IChart } from "../index";

const BarChart = (props: ChartProps) => {
  return <Chart {...props} />;
};

const definition: IChart = {
  type: "bar",
  component: BarChart,
};

export default definition;
