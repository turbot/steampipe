import Chart from "../Chart";
import { ChartProps, IChart } from "../index";

const PieChart = (props: ChartProps) => {
  return <Chart {...props} />;
};

const definition: IChart = {
  type: "pie",
  component: PieChart,
};

export default definition;
