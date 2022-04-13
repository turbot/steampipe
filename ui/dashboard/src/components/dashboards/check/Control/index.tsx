import CheckCard from "../common/CheckCard";
import LayoutPanel from "../../layout/common/LayoutPanel";
import LoadingIndicator from "../../LoadingIndicator";
import React, { useMemo } from "react";
import {
  AlarmIcon,
  InfoIcon,
  OKIcon,
  SortAscendingIcon,
  SortDescendingIcon,
} from "../../../../constants/icons";
import { CheckControl, CheckProps, CheckSummary } from "../common";
import { classNames } from "../../../../utils/styles";
import { useSortBy, useTable } from "react-table";

interface ControlsTableProps {
  loading: boolean;
  control: CheckControl | null;
}

const ControlsTable = ({ loading, control }: ControlsTableProps) => {
  const columns = useMemo(
    () => [
      { Header: "Status", accessor: "status", autoFit: true },
      { Header: "Reason", accessor: "reason" },
      { Header: "Resource", accessor: "resource" },
    ],
    []
  );
  const rowData = useMemo(
    () => (control && control.results ? control.results : []),
    [control]
  );
  const { getTableProps, getTableBodyProps, headerGroups, prepareRow, rows } =
    useTable({ columns, data: rowData }, useSortBy);
  return (
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
                  "px-4 py-3 text-left text-sm font-normal tracking-wider whitespace-nowrap",
                  column.autoFit ? "w-px" : ""
                )}
              >
                {column.render("Header")}
                {column.isSortedDesc ? (
                  <SortDescendingIcon
                    className={
                      column.isSorted ? "ml-1" : "ml-1 h-2 w-2 invisible"
                    }
                  />
                ) : (
                  <SortAscendingIcon
                    className={
                      column.isSorted ? "ml-1" : "ml-1 h-2 w-2 invisible"
                    }
                  />
                )}
              </th>
            ))}
          </tr>
        ))}
      </thead>
      <tbody {...getTableBodyProps()} className="divide-y divide-table-divide">
        {rows.map((row) => {
          prepareRow(row);
          return (
            <tr {...row.getRowProps()}>
              {row.cells.map((cell) => (
                <td
                  {...cell.getCellProps()}
                  className="px-4 py-4 align-top content-center text-sm whitespace-nowrap"
                >
                  {cell.column.Header === "Status" && (
                    <>
                      {["ALARM", "ERROR", "INVALID"].includes(
                        cell.value.toUpperCase()
                      ) && (
                        <div
                          className="whitespace-nowrap text-alert"
                          title={`Control in ${cell.value} status`}
                        >
                          <AlarmIcon />
                        </div>
                      )}
                      {cell.value.toUpperCase() === "OK" && (
                        <div
                          className="whitespace-nowrap text-ok"
                          title="Control in OK status"
                        >
                          <OKIcon />
                        </div>
                      )}
                      {cell.value.toUpperCase() === "INFO" && (
                        <div
                          className="whitespace-nowrap text-tbd"
                          title="Control in info status"
                        >
                          <InfoIcon />
                        </div>
                      )}
                    </>
                  )}
                  {cell.column.Header === "Reason" && (
                    <div className="text-black-scale-5 italic">
                      {cell.value}
                    </div>
                  )}
                  {cell.column.Header === "Resource" && cell.value}
                </td>
              ))}
            </tr>
          );
        })}
        {loading && (
          <tr>
            <td
              className="px-4 py-4 align-top content-center text-sm whitespace-nowrap italic text-black-scale-4"
              colSpan={3}
            >
              <LoadingIndicator className="mr-2" />
              Loading
            </td>
          </tr>
        )}
      </tbody>
    </table>
  );
};

const Control = (props: CheckProps) => {
  const { loading, summary, control } = useMemo(() => {
    const summary = props.root?.summary?.status;
    if (!summary) {
      return {
        loading: true,
        summary: {} as CheckSummary,
        control: null,
      };
    }
    const control = props.root?.controls?.[0];
    if (!control) {
      return {
        loading: true,
        summary: {} as CheckSummary,
        control: null,
      };
    }
    return { loading: false, summary, control };
  }, [props.root]);

  return (
    <LayoutPanel
      definition={{
        name: props.name,
        width: props.width,
      }}
    >
      <div className="col-span-12 grid grid-cols-5 gap-4">
        <CheckCard loading={loading} status="ok" value={summary.ok} />
        <CheckCard loading={loading} status="skip" value={summary.skip} />
        <CheckCard loading={loading} status="info" value={summary.info} />
        <CheckCard loading={loading} status="alarm" value={summary.alarm} />
        <CheckCard loading={loading} status="error" value={summary.error} />
      </div>
      <div className="col-span-12">
        <ControlsTable loading={loading} control={control} />
      </div>
    </LayoutPanel>
  );
};

export default Control;
