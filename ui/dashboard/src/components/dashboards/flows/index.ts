import Sankey from "./Sankey";
import Table from "../Table";
import { IFlow } from "./types";

const TableWrapper: IFlow = {
  type: "table",
  component: Table,
};

const flows = {
  [Sankey.type]: Sankey,
  [TableWrapper.type]: TableWrapper,
};

export { flows };
