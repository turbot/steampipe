import { getComponent } from "../index";
import { IGraph } from "./types";
const Table = getComponent("table");

const graphsMap = {};

const getGraphComponent = (key: string): IGraph => graphsMap[key];

const registerGraphComponent = (key: string, component: IGraph) => {
  graphsMap[key] = component;
};

const TableWrapper: IGraph = {
  type: "table",
  component: Table,
};

registerGraphComponent(TableWrapper.type, TableWrapper);

export { getGraphComponent, registerGraphComponent };
