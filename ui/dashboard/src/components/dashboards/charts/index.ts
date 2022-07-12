import AreaChart from "./AreaChart";
import BarChart from "./BarChart";
import ColumnChart from "./ColumnChart";
import DonutChart from "./DonutChart";
import LineChart from "./LineChart";
import PieChart from "./PieChart";
import Table from "../Table";
import { IChart } from "./types";

const TableWrapper: IChart = {
  type: "table",
  component: Table,
};

const charts = {
  [AreaChart.type]: AreaChart,
  [BarChart.type]: BarChart,
  [ColumnChart.type]: ColumnChart,
  [DonutChart.type]: DonutChart,
  [LineChart.type]: LineChart,
  [PieChart.type]: PieChart,
  [TableWrapper.type]: TableWrapper,
};

export default charts;
