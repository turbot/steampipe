import { getComponent } from "../index";
import { IInput } from "./types";
const Table = getComponent("table");

const flowsMap = {};

const getInputComponent = (key: string): IInput => flowsMap[key];

const registerInputComponent = (key: string, component: IInput) => {
  flowsMap[key] = component;
};

const TableWrapper: IInput = {
  type: "table",
  component: Table,
};

registerInputComponent(TableWrapper.type, TableWrapper);

export { getInputComponent, registerInputComponent };
