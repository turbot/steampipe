import {
  BasePrimitiveProps,
  ExecutablePrimitiveProps,
  isNumericCol,
  LeafNodeDataColumn,
  LeafNodeDataRow,
} from "../common";
import { classNames } from "../../../utils/styles";
import { getInterpolatedTemplateValue } from "../../../utils/template";
import { isObject } from "lodash";
import { Link } from "react-router-dom";
import { memo, useEffect, useMemo, useState } from "react";
import {
  SortAscendingIcon,
  SortDescendingIcon,
} from "../../../constants/icons";
import { usePanel } from "../../../hooks/usePanel";
import { useSortBy, useTable } from "react-table";

type TableColumnDisplay = "all" | "none";
type TableColumnWrap = "all" | "none";

interface TableColumnInfo {
  Header: string;
  accessor: string;
  name: string;
  data_type_name: string;
  wrap: TableColumnWrap;
  href_template?: string;
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
    let colHref: string | null = null;
    let colWrap: TableColumnWrap = "none";
    if (properties && properties.columns && properties.columns[col.name]) {
      const c = properties.columns[col.name];
      if (c.display === "none") {
        hiddenColumns.push(col.name);
      }
      if (c.wrap) {
        colWrap = c.wrap as TableColumnWrap;
      }
      if (c.href) {
        colHref = c.href;
      }
    }

    const colInfo: TableColumnInfo = {
      Header: col.name,
      accessor: col.name,
      name: col.name,
      data_type_name: col.data_type_name,
      wrap: colWrap,
    };
    if (colHref) {
      colInfo.href_template = colHref;
    }
    return colInfo;
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

interface TableRow {
  [key: string]: any;
}

interface CellValueProps {
  column: TableColumnInfo;
  row?: TableRow;
  value: any;
  showTitle?: boolean;
}

const CellValue = ({
  column,
  row,
  value,
  showTitle = false,
}: CellValueProps) => {
  const [href, setHref] = useState<string | null>(null);
  const [rendered, setRendered] = useState(false);

  // Calculate a link for this cell
  useEffect(() => {
    const render = async () => {
      const renderedVal = await getInterpolatedTemplateValue(
        column.href_template,
        { row }
      );
      // console.log(renderedVal);
      setHref(renderedVal);
      setRendered(true);
    };

    if (column.href_template) {
      render();
    } else {
      setHref(null);
      setRendered(true);
    }
  }, [column, row]);

  if (!rendered) {
    return null;
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
  if (dataType === "timestamp" || dataType === "timestamptz") {
    return <span className="tabular-nums">{value}</span>;
  }
  if (isNumericCol(dataType)) {
    return <span className="tabular-nums">{value}</span>;
  }
  // Fallback is just show it as a string
  return href ? (
    <Link to={href} className="link-highlight">
      {value}
    </Link>
  ) : (
    <span title={showTitle ? `${column.name}=${value}` : undefined}>
      {value}
    </span>
  );
};

const MemoCellValue = memo(CellValue);

interface TableColumnOptions {
  display?: TableColumnDisplay;
  href?: string;
  wrap?: TableColumnWrap;
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
  const { setZoomIconClassName } = usePanel();
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

  useEffect(() => {
    setZoomIconClassName("bg-background left-1 top-3");
  }, [setZoomIconClassName]);

  return props.data ? (
    <>
      <table
        {...getTableProps()}
        className="min-w-full divide-y divide-table-divide border-t border-table-border overflow-hidden"
      >
        <thead className="bg-table-head text-table-head">
          {headerGroups.map((headerGroup) => (
            <tr {...headerGroup.getHeaderGroupProps()}>
              {headerGroup.headers.map((column) => (
                <th
                  {...column.getHeaderProps(column.getSortByToggleProps())}
                  scope="col"
                  className={classNames(
                    "py-3 text-left text-sm font-normal tracking-wider whitespace-nowrap",
                    isNumericCol(column.data_type_name) ? " text-right" : "pl-4"
                  )}
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
              ))}
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
                {row.cells.map((cell) => (
                  <td
                    {...cell.getCellProps()}
                    className={classNames(
                      "px-4 py-4 align-top content-center text-sm",
                      isNumericCol(cell.column.data_type_name)
                        ? "text-right"
                        : "",
                      cell.column.wrap === "all"
                        ? "break-all"
                        : "whitespace-nowrap"
                    )}
                  >
                    <MemoCellValue
                      column={cell.column}
                      row={row.values}
                      value={cell.value}
                    />
                  </td>
                ))}
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
        const rowObj = {};
        (props.data?.columns || []).forEach((col, index) => {
          rowObj[col.name] = row[index];
        });
        return (
          <div key={rowIndex} className="space-y-2">
            {row.map((cellValue, columnIndex) => {
              const col = props.data?.columns[columnIndex];
              if (!col) {
                return null;
              }
              const columnOverrides =
                props.properties?.columns && props.properties.columns[col.name];
              if (columnOverrides?.display === "none") {
                return null;
              }
              const newColDef: TableColumnInfo = {
                ...col,
                Header: col.name,
                accessor: col.name,
                wrap: columnOverrides?.wrap ? columnOverrides.wrap : "none",
                href_template: columnOverrides?.href,
              };

              return (
                <div key={`${col?.name}-${rowIndex}`}>
                  <span className="block text-sm text-table-head truncate">
                    {col?.name}
                  </span>
                  <span
                    className={classNames(
                      "block",
                      newColDef.wrap === "all" ? "break-words" : "truncate"
                    )}
                  >
                    <MemoCellValue
                      column={newColDef}
                      row={rowObj}
                      value={cellValue}
                      showTitle
                    />
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
