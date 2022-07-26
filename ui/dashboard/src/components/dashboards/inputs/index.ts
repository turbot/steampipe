import { getComponent } from "../index";
import { IInput } from "./types";
const Table = getComponent("table");

const inputsMap = {};

const getInputComponent = (key: string): IInput => inputsMap[key];

const registerInputComponent = (key: string, component: IInput) => {
  inputsMap[key] = component;
};

const TableWrapper: IInput = {
  type: "table",
  component: Table,
};

registerInputComponent(TableWrapper.type, TableWrapper);

export { getInputComponent, registerInputComponent };
