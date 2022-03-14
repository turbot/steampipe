import ExternalLink from "../../ExternalLink";
import useDeepCompareEffect from "use-deep-compare-effect";
import {
  BasePrimitiveProps,
  ExecutablePrimitiveProps,
  isNumericCol,
  LeafNodeDataColumn,
  LeafNodeDataRow,
} from "../common";
import { classNames } from "../../../utils/styles";
import {
  getInterpolatedTemplateValue,
  RenderResults,
  renderTemplates,
} from "../../../utils/template";
import { isEmpty, isObject } from "lodash";
import { memo, useEffect, useMemo, useState } from "react";
import {
  SortAscendingIcon,
  SortDescendingIcon,
} from "../../../constants/icons";
import { useSortBy, useTable } from "react-table";

type TableColumnDisplay = "all" | "none";
type TableColumnWrap = "all" | "none";

interface TableColumnInfo {
  Header: string;
  accessor: string;
  name: string;
  data_type_name: string;
  display?: "all" | "none";
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

interface CellValueProps {
  column: TableColumnInfo;
  rowIndex: number;
  rowTemplateData: RenderResults[];
  value: any;
  showTitle?: boolean;
}

// // create a worker pool using an external worker script
// const jqRenderPool = createPool("../../../workers/renderJqTemplate", {
//   maxWorkers: 3,
// });

// const workers = [];
// function getWorker(url, metaUrl) {
//   var w;
//   if (workers.length > 0) {
//     w = workers.pop();
//   } else {
//     // @ts-ignore
//     // w = new Worker(
//     //   new URL("../../../workers/renderJqTemplate", import.meta.url)
//     // );
//     w = new Worker(url);
//   }
//   return w;
// }
//
// const releaseWorker = (worker) => {
//   // @ts-ignore
//   workers.push(worker);
// };
//
// function WorkerPool(url, metaUrl) {
//   // @ts-ignore
//   this.url = url;
//   // @ts-ignore
//   this.metaUrl = metaUrl;
//   // @ts-ignore
//   this.pool = [];
// }
// WorkerPool.prototype.getWorker = function () {
//   var w;
//   if (this.pool.length > 0) {
//     w = this.pool.pop();
//   } else {
//     // @ts-ignore
//     w = new Worker(new URL(this.url, this.metaUrl));
//   }
//   return w;
// };
// WorkerPool.prototype.releaseWorker = function (w) {
//   this.pool.push(w);
// };

// var pool = new WorkerPool("../../../workers/renderJqTemplate", import.meta.url);

const CellValue = ({
  column,
  rowIndex,
  rowTemplateData,
  value,
  showTitle = false,
}: CellValueProps) => {
  const [href, setHref] = useState<string | null>(null);

  // Calculate a link for this cell
  useEffect(() => {
    const renderedTemplateObj = rowTemplateData[rowIndex];
    if (!renderedTemplateObj) {
      setHref(null);
      return;
    }
    const renderedTemplateForColumn = renderedTemplateObj[column.name];
    if (!renderedTemplateForColumn) {
      setHref(null);
      return;
    }
    if (renderedTemplateForColumn.result) {
      setHref(renderedTemplateForColumn.result);
    }
  }, [rowIndex, rowTemplateData]);

  const dataType = column.data_type_name.toLowerCase();
  if (value === null || value === undefined) {
    return href ? (
      <ExternalLink
        to={href}
        className="link-highlight"
        title={showTitle ? `${column.name}=null` : undefined}
      >
        <>null</>
      </ExternalLink>
    ) : (
      <span
        className="text-foreground-lightest"
        title={showTitle ? `${column.name}=null` : undefined}
      >
        <>null</>
      </span>
    );
  }
  if (dataType === "bool") {
    // True should be
    return href ? (
      <ExternalLink
        to={href}
        className="link-highlight"
        title={showTitle ? `${column.name}=${value.toString()}` : undefined}
      >
        <>{value.toString()}</>
      </ExternalLink>
    ) : (
      <span
        className={classNames(value ? null : "text-foreground-light")}
        title={showTitle ? `${column.name}=${value.toString()}` : undefined}
      >
        <>{value.toString()}</>
      </span>
    );
  }
  if (dataType === "jsonb" || isObject(value)) {
    const asJsonString = JSON.stringify(value, null, 2);
    return href ? (
      <ExternalLink
        to={href}
        className="link-highlight"
        title={showTitle ? `${column.name}=${asJsonString}` : undefined}
      >
        <>{asJsonString}</>
      </ExternalLink>
    ) : (
      <span title={showTitle ? `${column.name}=${asJsonString}` : undefined}>
        {asJsonString}
      </span>
    );
  }
  if (dataType === "text") {
    if (value.match("^https?://")) {
      return (
        <ExternalLink
          className="link-highlight tabular-nums"
          to={value}
          title={showTitle ? `${column.name}=${value}` : undefined}
        >
          {value}
        </ExternalLink>
      );
    }
    const mdMatch = value.match("^\\[(.*)\\]\\((https?://.*)\\)$");
    if (mdMatch) {
      return (
        <ExternalLink
          className="tabular-nums"
          to={mdMatch[2]}
          title={showTitle ? `${column.name}=${value}` : undefined}
        >
          {mdMatch[1]}
        </ExternalLink>
      );
    }
  }
  if (dataType === "timestamp" || dataType === "timestamptz") {
    return href ? (
      <ExternalLink
        to={href}
        className="link-highlight tabular-nums"
        title={showTitle ? `${column.name}=${value}` : undefined}
      >
        {value}
      </ExternalLink>
    ) : (
      <span
        className="tabular-nums"
        title={showTitle ? `${column.name}=${value}` : undefined}
      >
        {value}
      </span>
    );
  }
  if (isNumericCol(dataType)) {
    return href ? (
      <ExternalLink
        to={href}
        className="link-highlight tabular-nums"
        title={showTitle ? `${column.name}=${value}` : undefined}
      >
        {value}
      </ExternalLink>
    ) : (
      <span
        className="tabular-nums"
        title={showTitle ? `${column.name}=${value}` : undefined}
      >
        {value}
      </span>
    );
  }
  // Fallback is just show it as a string
  return href ? (
    <ExternalLink
      to={href}
      className="link-highlight tabular-nums"
      title={showTitle ? `${column.name}=${value}` : undefined}
    >
      {value}
    </ExternalLink>
  ) : (
    <span
      className="tabular-nums"
      title={showTitle ? `${column.name}=${value}` : undefined}
    >
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
  const [rowTemplateData, setRowTemplateData] = useState<RenderResults[]>([]);
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

  useDeepCompareEffect(() => {
    const doRender = async () => {
      const templates = Object.fromEntries(
        columns
          .filter((col) => col.display !== "none" && !!col.href_template)
          .map((col) => [col.name, col.href_template as string])
      );
      if (isEmpty(templates)) {
        setRowTemplateData([]);
        return;
      }
      const data = rows.map((row) => row.values);
      const renderedResults = await renderTemplates(templates, data);
      setRowTemplateData(renderedResults);
    };

    if (columns.length === 0 || rows.length === 0) {
      setRowTemplateData([]);
      return;
    }

    doRender();
  }, [columns, rows]);

  return props.data ? (
    <>
      <table
        {...getTableProps()}
        className={classNames(
          "min-w-full divide-y divide-table-divide overflow-hidden",
          props.title ? "border-t border-table-divide" : null
        )}
      >
        <thead className="bg-table-head text-table-head">
          {headerGroups.map((headerGroup) => (
            <tr {...headerGroup.getHeaderGroupProps()}>
              {headerGroup.headers.map((column) => (
                <th
                  {...column.getHeaderProps(column.getSortByToggleProps())}
                  scope="col"
                  className={classNames(
                    "py-3 text-left text-sm font-normal tracking-wider whitespace-nowrap pl-4",
                    isNumericCol(column.data_type_name) ? "text-right" : null
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
          {rows.map((row, index) => {
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
                      rowIndex={index}
                      rowTemplateData={rowTemplateData}
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

interface LineModeRow {
  [key: string]: any;
}

interface LineModeRows {
  row: LeafNodeDataRow;
  obj: LineModeRow;
}

const LineView = (props: TableProps) => {
  const [columns, setColumns] = useState<TableColumnInfo[]>([]);
  const [rows, setRows] = useState<LineModeRows[]>([]);
  const [rowTemplateData, setRowTemplateData] = useState<RenderResults[]>([]);

  useEffect(() => {
    if (!props.data || !props.data.columns || !props.data.rows) {
      return;
    }
    const newColumns: TableColumnInfo[] = [];
    props.data.columns.forEach((col) => {
      const columnOverrides =
        props.properties?.columns && props.properties.columns[col.name];
      const newColDef: TableColumnInfo = {
        ...col,
        Header: col.name,
        accessor: col.name,
        display: columnOverrides?.display ? columnOverrides.display : "all",
        wrap: columnOverrides?.wrap ? columnOverrides.wrap : "none",
        href_template: columnOverrides?.href,
      };
      newColumns.push(newColDef);
    });

    const newRows: LineModeRows[] = [];
    props.data.rows.forEach((row) => {
      const rowObj = {};
      newColumns.forEach((col, index) => {
        rowObj[col.name] = row[index];
      });
    });

    setColumns(newColumns);
    setRows(newRows);
  }, [props.data]);

  useDeepCompareEffect(() => {
    const doRender = async () => {
      const templates = Object.fromEntries(
        columns
          .filter((col) => col.display !== "none" && !!col.href_template)
          .map((col) => [col.name, col.href_template as string])
      );
      if (isEmpty(templates)) {
        setRowTemplateData([]);
        return;
      }
      const data = rows.map((row) => row.obj);
      const renderedResults = await renderTemplates(templates, data);
      console.log(renderedResults);
      setRowTemplateData(renderedResults);
    };

    if (columns.length === 0 || rows.length === 0) {
      setRowTemplateData([]);
      return;
    }

    doRender();
  }, [columns, rows]);

  if (columns.length === 0 || rows.length === 0) {
    return null;
  }

  return (
    <div className="px-4 py-3 space-y-4">
      {rows.map((rowInfo, rowIndex) => {
        return (
          <div key={rowIndex} className="space-y-2">
            {rowInfo.row.map((cellValue, columnIndex) => {
              const col = columns[columnIndex];
              if (!col || col.display === "none") {
                return null;
              }
              return (
                <div key={`${col.name}-${rowIndex}`}>
                  <span className="block text-sm text-table-head truncate">
                    {col.name}
                  </span>
                  <span
                    className={classNames(
                      "block",
                      col.wrap === "all" ? "break-words" : "truncate"
                    )}
                  >
                    <MemoCellValue
                      column={col}
                      rowIndex={rowIndex}
                      rowTemplateData={rowTemplateData}
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
