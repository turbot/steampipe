import { getComponent } from "../index";
import { IHierarchy } from "./types";
const Table = getComponent("table");

const flowsMap = {};

const getHierarchyComponent = (key: string): IHierarchy => flowsMap[key];

const registerHierarchyComponent = (key: string, component: IHierarchy) => {
  flowsMap[key] = component;
};

const TableWrapper: IHierarchy = {
  type: "table",
  component: Table,
};

registerHierarchyComponent(TableWrapper.type, TableWrapper);

export { getHierarchyComponent, registerHierarchyComponent };
