import Table from "../Table";
import Tree from "./Tree";
import { IHierarchy } from "./types";

const TableWrapper: IHierarchy = {
  type: "table",
  component: Table,
};

const hierarchies = {
  [TableWrapper.type]: TableWrapper,
  [Tree.type]: Tree,
};

export { hierarchies };
