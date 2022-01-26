import Icon from "../../Icon";
import {
  BasePrimitiveProps,
  ExecutablePrimitiveProps,
  LeafNodeDataColumn,
  LeafNodeDataItem,
} from "../common";
import {
  sortAscendingIcon,
  sortDescendingIcon,
} from "../../../constants/icons";
import { useMemo } from "react";
import { useSortBy, useTable } from "react-table";

interface ColumnInfo {
  Header: string;
  accessor: string;
}

const getColumns = (columns: LeafNodeDataColumn[]) => {
  if (!columns || columns.length === 0) {
    return [];
  }
  return columns.map((col) => ({ Header: col.name, accessor: col.name }));
};

const getData = (columns: ColumnInfo[], items: LeafNodeDataItem) => {
  if (!columns || columns.length === 0) {
    return [];
  }

  if (!items || items.length === 0) {
    return [];
  }
  return items;
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

const CellValue = ({ value }) => {
  if (value === undefined || value === null) {
    return <span className="text-black-scale-3">null</span>;
  }
  return value;
};

export type TableProps = BasePrimitiveProps & ExecutablePrimitiveProps;

// TODO retain full width on mobile, no padding
const Table = (props: TableProps) => {
  const columns = useMemo(
    () => getColumns(props.data ? props.data.columns : []),
    [props.data]
  );
  const rowData = useMemo(
    () => getData(columns, props.data ? props.data.items : []),
    [columns, props.data]
  );
  const { getTableProps, getTableBodyProps, headerGroups, prepareRow, rows } =
    useTable({ columns, data: rowData }, useSortBy);
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
                  className="px-4 py-3 text-left text-sm font-normal tracking-wider whitespace-nowrap"
                >
                  {column.render("Header")}
                  <Icon
                    className={column.isSorted ? "ml-1" : "ml-1 invisible"}
                    icon={
                      column.isSortedDesc
                        ? sortDescendingIcon
                        : sortAscendingIcon
                    }
                    onClick={undefined}
                    rotation={undefined}
                  />
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
                {row.cells.map((cell) => {
                  return (
                    <td
                      {...cell.getCellProps()}
                      className="px-4 py-4 align-top content-center text-sm whitespace-nowrap"
                    >
                      <CellValue value={cell.value} />
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
