import { getComponent } from "../index";
import { IChart } from "./types";
const Table = getComponent("table");

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
