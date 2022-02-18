import {
  BasePrimitiveProps,
  ExecutablePrimitiveProps,
  LeafNodeDataColumn,
  LeafNodeDataRow,
} from "../common";
import { classNames } from "../../../utils/styles";
import { isObject } from "lodash";
import {
  SortAscendingIcon,
  SortDescendingIcon,
} from "../../../constants/icons";
import { useMemo } from "react";
import { useSortBy, useTable } from "react-table";

type TableColumnWrap = "all" | "none";

interface TableColumnInfo {
  Header: string;
  accessor: string;
  wrap?: TableColumnWrap;
}

const getColumns = (
  cols: LeafNodeDataColumn[],
  properties?: TableProperties
): { columns: TableColumnInfo[]; hiddenColumns: string[] } => {
  if (!cols || cols.length === 0) {
    return { columns: [], hiddenColumns: [] };
  }

  const hiddenColumns: string[] = [];
  const columns: TableColumnInfo[] = cols.map((col) => {
    let colWrap: TableColumnWrap = "none";
    if (properties && properties.columns && properties.columns[col.name]) {
      const c = properties.columns[col.name];
      if (c.display === "none") {
        hiddenColumns.push(col.name);
      }
      if (c.wrap) {
        colWrap = c.wrap as TableColumnWrap;
      }
    }

    return {
      Header: col.name,
      accessor: col.name,
      data_type_name: col.data_type_name,
      wrap: colWrap,
    };
  });
  return { columns, hiddenColumns };
};

const getData = (columns: TableColumnInfo[], rows: LeafNodeDataRow) => {
  if (!columns || columns.length === 0) {
    return [];
  }

  if (!rows || rows.length === 0) {
    return [];
  }
  return rows.map((r) => {
    const rowData = {};
    for (let colIndex = 0; colIndex < r.length; colIndex++) {
      rowData[columns[colIndex].accessor] = r[colIndex];
    }
    return rowData;
  });
};

interface CellValueProps {
  column?: LeafNodeDataColumn;
  value: any;
  showTitle?: boolean;
}

const CellValue = ({ column, value, showTitle = false }: CellValueProps) => {
  if (!column) {
    return <>null</>;
  }

  const dataType = column.data_type_name.toLowerCase();
  if (value === null || value === undefined) {
    return (
      <span
        className="text-black-scale-3"
        title={showTitle ? `${column.name}=null` : undefined}
      >
        null
      </span>
    );
  }
  if (dataType === "bool") {
    // True should be
    return (
      <span
        className={classNames(value ? null : "text-foreground-light")}
        title={showTitle ? `${column.name}=${value.toString()}` : undefined}
      >
        {value.toString()}
      </span>
    );
  }
  if (dataType === "jsonb" || isObject(value)) {
    const asJsonString = JSON.stringify(value, null, 2);
    return (
      <span title={showTitle ? `${column.name}=${asJsonString}` : undefined}>
        {asJsonString}
      </span>
    );
  }
  if (dataType === "text") {
    if (value.match("^https?://")) {
      return (
        <a
          className="text-link"
          target="_blank"
          rel="noopener noreferrer"
          href={value}
          title={showTitle ? `${column.name}=${value}` : undefined}
        >
          {value}
        </a>
      );
    }
    const mdMatch = value.match("^\\[(.*)\\]\\((https?://.*)\\)$");
    if (mdMatch) {
      return (
        <a
          target="_blank"
          rel="noopener noreferrer"
          href={mdMatch[2]}
          title={showTitle ? `${column.name}=${value}` : undefined}
        >
          {mdMatch[1]}
        </a>
      );
    }
  }
  // Fallback is just show it as a string
  return (
    <span title={showTitle ? `${column.name}=${value}` : undefined}>
      {value}
    </span>
  );
};

interface TableColumnOptions {
  display?: string;
  wrap?: string;
}

type TableColumns = {
  [column: string]: TableColumnOptions;
};

type TableType = "table" | "line" | null;

export type TableProperties = {
  type?: TableType;
  columns?: TableColumns;
};

export type BaseTableProps = BasePrimitiveProps & ExecutablePrimitiveProps;

export type TableProps = BaseTableProps & {
  properties?: TableProperties;
};

// TODO retain full width on mobile, no padding
const TableView = (props: TableProps) => {
  const { columns, hiddenColumns } = useMemo(
    () => getColumns(props.data ? props.data.columns : [], props.properties),
    [props.data, props.properties]
  );
  const rowData = useMemo(
    () => getData(columns, props.data ? props.data.rows : []),
    [columns, props.data]
  );
  const { getTableProps, getTableBodyProps, headerGroups, prepareRow, rows } =
    useTable(
      { columns, data: rowData, initialState: { hiddenColumns } },
      useSortBy
    );
  return props.data ? (
    <>
      <table
        {...getTableProps()}
        className="min-w-full divide-y divide-table-divide border-t border-table-border overflow-hidden"
      >
        <thead className="bg-table-head text-table-head">
          {headerGroups.map((headerGroup) => (
            <tr {...headerGroup.getHeaderGroupProps()}>
              {headerGroup.headers.map((column) => {
                return (
                  <th
                    {...column.getHeaderProps(column.getSortByToggleProps())}
                    scope="col"
                    className="px-4 py-3 text-left text-sm font-normal tracking-wider whitespace-nowrap"
                  >
                    {column.render("Header")}
                    {column.isSortedDesc ? (
                      <SortDescendingIcon className="inline-block h-4 w-4" />
                    ) : (
                      <SortAscendingIcon
                        className={classNames(
                          "inline-block h-4 w-4",
                          !column.isSorted ? "invisible" : null
                        )}
                      />
                    )}
                  </th>
                );
              })}
            </tr>
          ))}
        </thead>
        <tbody
          {...getTableBodyProps()}
          className="divide-y divide-table-divide"
        >
          {props.data && rows.length === 0 && (
            <tr>
              <td
                className="px-4 py-4 align-top content-center text-sm italic whitespace-nowrap"
                colSpan={columns.length}
              >
                No results
              </td>
            </tr>
          )}
          {rows.map((row) => {
            prepareRow(row);
            return (
              <tr {...row.getRowProps()}>
                {row.cells.map((cell, index) => {
                  return (
                    <td
                      {...cell.getCellProps()}
                      className={classNames(
                        "px-4 py-4 align-top content-center text-sm",
                        cell.column.wrap === "all"
                          ? "break-all"
                          : "whitespace-nowrap"
                      )}
                    >
                      <CellValue
                        column={props.data?.columns[index]}
                        value={cell.value}
                      />
                    </td>
                  );
                })}
              </tr>
            );
          })}
        </tbody>
      </table>
    </>
  ) : null;
};

const LineView = (props: TableProps) => {
  if (
    !props.data ||
    !props.data.columns ||
    props.data.columns.length === 0 ||
    !props.data.rows ||
    props.data.rows.length === 0
  ) {
    return null;
  }

  return (
    <div className="space-y-4">
      {props.data.rows.map((row, rowIndex) => {
        return (
          <div key={rowIndex} className="space-y-2">
            {row.map((cellValue, columnIndex) => {
              const col = props.data?.columns[columnIndex];
              return (
                <div key={`${col?.name}-${rowIndex}`}>
                  <span className="block text-sm text-table-head truncate">
                    {col?.name}
                  </span>
                  <span className="block truncate">
                    <CellValue column={col} value={cellValue} showTitle />
                  </span>
                </div>
              );
            })}
          </div>
        );
      })}
    </div>
  );
};

const Table = (props: TableProps) => {
  if (props.properties && props.properties.type === "line") {
    return <LineView {...props} />;
  }
  return <TableView {...props} />;
};

export default Table;
