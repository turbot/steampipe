import Chart from "../Chart";
import { ChartProps, IChart } from "../types";

const LineChart = (props: ChartProps) => {
  return <Chart {...props} />;
};

const definition: IChart = {
  type: "line",
  component: LineChart,
};

export default definition;
