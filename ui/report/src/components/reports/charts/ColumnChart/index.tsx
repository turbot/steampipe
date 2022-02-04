import Chart from "../Chart";
import { ChartProps, IChart } from "../index";

const ColumnChart = (props: ChartProps) => {
  return <Chart {...props} />;
};

const definition: IChart = {
  type: "column",
  component: ColumnChart,
};

export default definition;
