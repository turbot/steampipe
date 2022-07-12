import Chart from "../Chart";
import { ChartProps, IChart } from "../types";

const AreaChart = (props: ChartProps) => {
  return <Chart {...props} />;
};

const definition: IChart = {
  type: "area",
  component: AreaChart,
};

export default definition;
