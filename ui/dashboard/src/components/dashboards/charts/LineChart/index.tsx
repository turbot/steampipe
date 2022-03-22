import Chart from "../Chart";
import { ChartProps, IChart } from "../index";

const LineChart = (props: ChartProps) => {
  return <Chart {...props} />;
};

const definition: IChart = {
  type: "line",
  component: LineChart,
};

export default definition;
