import Chart from "../Chart";
import { ChartProps, IChart } from "../types";

const PieChart = (props: ChartProps) => {
  return <Chart {...props} />;
};

const definition: IChart = {
  type: "pie",
  component: PieChart,
};

export default definition;
