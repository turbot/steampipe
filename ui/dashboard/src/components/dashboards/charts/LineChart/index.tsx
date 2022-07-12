import Chart from "../Chart";
import { ChartProps, IChart } from "../types";
import { registerChartComponent } from "../index";

const LineChart = (props: ChartProps) => {
  return <Chart {...props} />;
};

const definition: IChart = {
  type: "line",
  component: LineChart,
};

registerChartComponent(definition.type, definition);

export default definition;
