import { LeafNodeDataColumn } from "../components/dashboards/common";

const getColumn = (
  columns: LeafNodeDataColumn[],
  name: string
): LeafNodeDataColumn | undefined => {
  if (!columns || !name) {
    return undefined;
  }

  return columns.find((col) => col.name === name);
};

export { getColumn };
