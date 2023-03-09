import Chart from "../Chart";
import { ChartProps, IChart } from "../types";
import { registerChartComponent } from "../index";

const BarChart = (props: ChartProps) => {
  return <Chart {...props} />;
};

const definition: IChart = {
  type: "bar",
  component: BarChart,
};

registerChartComponent(definition.type, definition);

export default definition;
