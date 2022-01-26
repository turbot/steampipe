import { LeafNodeDataColumn } from "../components/reports/common";

const getColumnIndex = (
  columns: LeafNodeDataColumn[],
  name: string
): number => {
  if (!columns || !name) {
    return -1;
  }

  return columns.findIndex((col) => col.name === name);
};

const hasColumn = (columns: LeafNodeDataColumn[], name: string): boolean => {
  if (!columns || !name) {
    return false;
  }

  return getColumnIndex(columns, name) >= 0;
};

export { hasColumn };
