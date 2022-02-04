import Chart from "../Chart";
import { ChartProps, IChart } from "../index";

const DonutChart = (props: ChartProps) => {
  return <Chart {...props} />;
};

const definition: IChart = {
  type: "donut",
  component: DonutChart,
};

export default definition;
