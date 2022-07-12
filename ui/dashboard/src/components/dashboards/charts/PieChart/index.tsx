import Chart from "../Chart";
import { ChartProps, IChart } from "../types";
import { registerChartComponent } from "../index";

const PieChart = (props: ChartProps) => {
  return <Chart {...props} />;
};

const definition: IChart = {
  type: "pie",
  component: PieChart,
};

registerChartComponent(definition.type, definition);

export default definition;
