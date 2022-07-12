import { getComponent } from "../index";
import { IFlow } from "./types";
const Table = getComponent("table");

const flowsMap = {};

const getFlowComponent = (key: string): IFlow => flowsMap[key];

const registerFlowComponent = (key: string, component: IFlow) => {
  flowsMap[key] = component;
};

const TableWrapper: IFlow = {
  type: "table",
  component: Table,
};

registerFlowComponent(TableWrapper.type, TableWrapper);

export { getFlowComponent, registerFlowComponent };
