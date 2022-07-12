import { getComponent } from "../index";
// import AreaChart from "./AreaChart";
// import BarChart from "./BarChart";
// import ColumnChart from "./ColumnChart";
// import DonutChart from "./DonutChart";
// import LineChart from "./LineChart";
// import PieChart from "./PieChart";
import { IChart } from "./types";
const Table = getComponent("table");

// const charts = {}
//   [AreaChart.type]: AreaChart,
//   [BarChart.type]: BarChart,
//   [ColumnChart.type]: ColumnChart,
//   [DonutChart.type]: DonutChart,
//   [LineChart.type]: LineChart,
//   [PieChart.type]: PieChart,
//   [TableWrapper.type]: TableWrapper,
// };

const chartsMap = {};

const getChartComponent = (key: string): IChart => chartsMap[key];

const registerChartComponent = (key: string, component: IChart) => {
  chartsMap[key] = component;
};

const TableWrapper: IChart = {
  type: "table",
  component: Table,
};

registerChartComponent(TableWrapper.type, TableWrapper);

export { getChartComponent, registerChartComponent };
