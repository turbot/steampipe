import { getComponent } from "../index";
import { IHierarchy } from "./types";
const Table = getComponent("table");

const hierarchiesMap = {};

const getHierarchyComponent = (key: string): IHierarchy => hierarchiesMap[key];

const registerHierarchyComponent = (key: string, component: IHierarchy) => {
  hierarchiesMap[key] = component;
};

const TableWrapper: IHierarchy = {
  type: "table",
  component: Table,
};

registerHierarchyComponent(TableWrapper.type, TableWrapper);

export { getHierarchyComponent, registerHierarchyComponent };
