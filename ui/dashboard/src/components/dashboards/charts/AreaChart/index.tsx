import Chart from "../Chart";
import { ChartProps, IChart } from "../types";
import { registerChartComponent } from "../index";

const AreaChart = (props: ChartProps) => {
  return <Chart {...props} />;
};

const definition: IChart = {
  type: "area",
  component: AreaChart,
};

registerChartComponent(definition.type, definition);

export default definition;
