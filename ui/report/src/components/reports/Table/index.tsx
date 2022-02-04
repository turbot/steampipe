import moment from "moment";
import {
  BasePrimitiveProps,
  ExecutablePrimitiveProps,
  LeafNodeDataColumn,
  LeafNodeDataRow,
} from "../common";
import { classNames } from "../../../utils/styles";
import {
  SortAscendingIcon,
  SortDescendingIcon,
} from "../../../constants/icons";
import { useMemo } from "react";
import { useSortBy, useTable } from "react-table";

interface ColumnInfo {
  Header: string;
  accessor: string;
}

const getColumns = (
  cols: LeafNodeDataColumn[],
  properties?: TableProperties
) => {
  if (!cols || cols.length === 0) {
    return { columns: [], hiddenColumns: [] };
  }

  const hiddenColumns: string[] = [];
  const columns = cols.map((col) => {
    let colWrap = "none";
    if (properties && properties.columns && properties.columns[col.name]) {
      const c = properties.columns[col.name];
      if (c.display === "none") {
        hiddenColumns.push(col.name);
      }
      if (c.wrap) {
        colWrap = c.wrap;
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

const getData = (columns: ColumnInfo[], rows: LeafNodeDataRow) => {
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

// const TablePaging = ({
//   canNextPage,
//   canPreviousPage,
//   gotoPage,
//   nextPage,
//   pageCount,
//   pageIndex,
//   pageOptions,
//   pageSize,
//   previousPage,
//   setPageSize,
// }) => {
//   if (pageCount <= 1) {
//     return null;
//   }
//
//   return (
//     <div className="min-w-full px-6 py-4 border-t">
//       <button
//         className={
//           canPreviousPage
//             ? "mr-1 text-gray-600 hover:text-gray-800 text-sm"
//             : "mr-1 text-gray-300 text-sm cursor-default"
//         }
//         onClick={() => gotoPage(0)}
//         disabled={!canPreviousPage}
//       >
//         <Icon icon={firstPageIcon} />
//       </button>
//       <button
//         className={
//           canPreviousPage
//             ? "mr-2 text-gray-600 hover:text-gray-800 text-sm"
//             : "mr-2 text-gray-300 text-sm cursor-default"
//         }
//         onClick={() => previousPage()}
//         disabled={!canPreviousPage}
//       >
//         <Icon icon={previousPageIcon} />
//       </button>
//       <button
//         className={
//           canNextPage
//             ? "mr-1 text-gray-600 hover:text-gray-800 text-sm"
//             : "mr-1 text-gray-300 text-sm cursor-default"
//         }
//         onClick={() => nextPage()}
//         disabled={!canNextPage}
//       >
//         <Icon icon={nextPageIcon} />
//       </button>
//       <button
//         className={
//           canNextPage
//             ? "mr-2 text-gray-600 hover:text-gray-800 text-sm"
//             : "mr-2 text-gray-300 text-sm cursor-default"
//         }
//         onClick={() => gotoPage(pageCount - 1)}
//         disabled={!canNextPage}
//       >
//         <Icon icon={lastPageIcon} />
//       </button>
//       <span className="mr-1 text-xs">
//         Page {pageIndex + 1} of {pageOptions.length}.
//       </span>
//       <span className="mr-1 text-xs">
//         <span className={pageCount <= 1 ? "text-gray-300" : null}>
//           Go to page:
//         </span>{" "}
//         <input
//           className="w-20 text-xs border-gray-400 rounded-sm disabled:text-gray-300 disabled:border-gray-300"
//           type="number"
//           defaultValue={pageIndex + 1}
//           disabled={pageCount <= 1}
//           onChange={(e) => {
//             const page = e.target.value ? Number(e.target.value) - 1 : 0;
//             gotoPage(page);
//           }}
//         />
//       </span>
//       <select
//         className="text-xs border-gray-400 rounded-sm disabled:text-gray-300 disabled:border-gray-300"
//         disabled={pageCount <= 1}
//         value={pageSize}
//         onChange={(e) => {
//           setPageSize(Number(e.target.value));
//         }}
//       >
//         {[10, 20, 50, 100].map((pageSize) => (
//           <option key={pageSize} value={pageSize}>
//             Show {pageSize}
//           </option>
//         ))}
//       </select>
//     </div>
//   );
// };

const CellValue = ({ column, value }) => {
  const dataType = column.data_type_name.toLowerCase();
  if (value === null || value === undefined) {
    return <span className="text-black-scale-3">null</span>;
  }
  if (dataType === "bool") {
    return (
      <span className={classNames(value ? "" : "text-foreground-light")}>
        {value.toString()}
      </span>
    );
  }
  if (dataType === "jsonb") {
    return <>{JSON.stringify(value, null, 2)}</>;
  }
  if (dataType === "timestamptz") {
    return <>{moment(value).format()}</>;
  }
  if (dataType === "text") {
    if (value.match("^https?://")) {
      return (
        <a
          className="text-link"
          target="_blank"
          rel="noopener noreferrer"
          href={value}
        >
          {value}
        </a>
      );
    }
    const mdMatch = value.match("^\\[(.*)\\]\\((https?://.*)\\)$");
    if (mdMatch) {
      return (
        <a target="_blank" rel="noopener noreferrer" href={mdMatch[2]}>
          {mdMatch[1]}
        </a>
      );
    }
    // return <span style={{ color: "#ffbb1b" }}>{value}</span>;
    // return <span style={{ color: "#779fc8" }}>{value}</span>;
    return <span>{value}</span>;
  }
  return <>{JSON.stringify(value)}</>;
};

// const CellValue = ({ value }) => {
//   if (value === undefined || value === null) {
//     return <span className="text-black-scale-3">null</span>;
//   }
//   return value;
// };

interface TableColumnOptions {
  display?: string;
  wrap?: string;
}

type TableColumns = {
  [column: string]: TableColumnOptions;
};

export type TableProperties = {
  columns?: TableColumns;
};

export type BaseTableProps = BasePrimitiveProps & ExecutablePrimitiveProps;

export type TableProps = BaseTableProps & {
  properties?: TableProperties;
};

// TODO retain full width on mobile, no padding
const Table = (props: TableProps) => {
  const { columns, hiddenColumns } = useMemo(
    () => getColumns(props.data ? props.data.columns : [], props.properties),
    [props.data]
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
                        cell.column.wrap === "always"
                          ? "break-all"
                          : "whitespace-nowrap"
                      )}
                    >
                      <CellValue column={columns[index]} value={cell.value} />
                    </td>
                  );
                })}
              </tr>
            );
          })}
        </tbody>
      </table>
      {/*<TablePaging*/}
      {/*  canNextPage={canNextPage}*/}
      {/*  canPreviousPage={canPreviousPage}*/}
      {/*  gotoPage={gotoPage}*/}
      {/*  nextPage={nextPage}*/}
      {/*  pageCount={pageCount}*/}
      {/*  pageIndex={pageIndex}*/}
      {/*  pageOptions={pageOptions}*/}
      {/*  pageSize={pageSize}*/}
      {/*  previousPage={previousPage}*/}
      {/*  setPageSize={setPageSize}*/}
      {/*/>*/}
    </>
  ) : null;
};

export default Table;
