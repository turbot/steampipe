import {
  LeafNodeData,
  LeafNodeDataColumn,
} from "../components/dashboards/common";

const hasData = (data: LeafNodeData | undefined) => {
  return (
    !!data &&
    data.columns &&
    data.rows &&
    data.columns.length > 0 &&
    data.rows.length > 0
  );
};

const getColumn = (
  columns: LeafNodeDataColumn[],
  name: string,
): LeafNodeDataColumn | undefined => {
  if (!columns || !name) {
    return undefined;
  }

  return columns.find((col) => col.name === name);
};

export { getColumn, hasData };
