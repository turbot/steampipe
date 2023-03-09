import Chart from "../Chart";
import { ChartProps, IChart } from "../types";
import { registerChartComponent } from "../index";

const DonutChart = (props: ChartProps) => {
  return <Chart {...props} />;
};

const definition: IChart = {
  type: "donut",
  component: DonutChart,
};

registerChartComponent(definition.type, definition);

export default definition;
