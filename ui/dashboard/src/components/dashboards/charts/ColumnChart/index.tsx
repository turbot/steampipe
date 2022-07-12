import Chart from "../Chart";
import { ChartProps, IChart } from "../types";
import { registerChartComponent } from "../index";

const ColumnChart = (props: ChartProps) => {
  return <Chart {...props} />;
};

const definition: IChart = {
  type: "column",
  component: ColumnChart,
};

registerChartComponent(definition.type, definition);

export default definition;
