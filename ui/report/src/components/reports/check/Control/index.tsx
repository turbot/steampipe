import CheckCounter from "../common/CheckCounter";
import Icon from "../../../Icon";
import LayoutPanel from "../../layout/common/LayoutPanel";
import LoadingIndicator from "../../LoadingIndicator";
import React, { useMemo } from "react";
import {
  alarmIcon,
  infoIcon,
  okIcon,
  sortAscendingIcon,
  sortDescendingIcon,
} from "../../../../constants/icons";
import {
  CheckLeafNodeDataControl,
  CheckLeafNodeDataGroupSummaryStatus,
  CheckProps,
} from "../common";
import { classNames } from "../../../../utils/styles";
import { useSortBy, useTable } from "react-table";

interface ControlsTableProps {
  loading: boolean;
  control: CheckLeafNodeDataControl | null;
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
    [columns, control]
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
                <Icon
                  className={column.isSorted ? "ml-1" : "ml-1 invisible"}
                  icon={
                    column.isSortedDesc ? sortDescendingIcon : sortAscendingIcon
                  }
                />
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
              {row.cells.map((cell) => {
                console.log(cell.column.Header, cell.value);
                return (
                  <td
                    {...cell.getCellProps()}
                    className="px-4 py-4 align-top content-center text-sm whitespace-nowrap"
                  >
                    {cell.column.Header === "Status" && (
                      <>
                        {["ALARM", "ERROR", "INVALID"].includes(
                          cell.value.toUpperCase()
                        ) && (
                          <div className="whitespace-nowrap text-alert">
                            <Icon
                              icon={alarmIcon}
                              title={`Control in ${cell.value} status`}
                            />
                          </div>
                        )}
                        {cell.value.toUpperCase() === "OK" && (
                          <div className="whitespace-nowrap text-ok">
                            <Icon icon={okIcon} title="Control in OK status" />
                          </div>
                        )}
                        {cell.value.toUpperCase() === "INFO" && (
                          <div className="whitespace-nowrap text-tbd">
                            <Icon
                              icon={infoIcon}
                              title="Control in info status"
                            />
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
                );
              })}
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
    const summary = props.execution_tree?.root?.summary?.status;
    if (!summary) {
      return {
        loading: true,
        summary: {} as CheckLeafNodeDataGroupSummaryStatus,
        control: null,
      };
    }
    const control = props.execution_tree?.root?.controls?.[0];
    if (!control) {
      return {
        loading: true,
        summary: {} as CheckLeafNodeDataGroupSummaryStatus,
        control: null,
      };
    }
    return { loading: false, summary, control };
  }, [props.execution_tree]);

  return (
    <LayoutPanel
      definition={{
        name: props.name,
        width: props.width,
      }}
    >
      <div className="col-span-12 grid grid-cols-5 gap-4">
        <CheckCounter loading={loading} status="ok" value={summary.ok} />
        <CheckCounter loading={loading} status="skip" value={summary.skip} />
        <CheckCounter loading={loading} status="info" value={summary.info} />
        <CheckCounter loading={loading} status="alarm" value={summary.alarm} />
        <CheckCounter loading={loading} status="error" value={summary.error} />
      </div>
      <div className="col-span-12">
        <ControlsTable loading={loading} control={control} />
      </div>
    </LayoutPanel>
  );

  // return (
  //   <Children
  //     children={[
  //       {
  //         name: `${props.name}.container.wrapper`,
  //         node_type: "container",
  //         children: [
  //           {
  //             name: `${props.name}.container.summary_wrapper`,
  //             node_type: "container",
  //             children: [
  //               {
  //                 name: `${props.name}.counter.ok`,
  //                 node_type: "counter",
  //                 properties: { style: "ok" },
  //                 data: {
  //                   columns: [{ name: "OK", data_type_name: "INT8" }],
  //                   items: [{ OK: summary.ok }],
  //                 },
  //               },
  //               {
  //                 name: `${props.name}.counter.skip`,
  //                 node_type: "counter",
  //                 properties: { style: "plain" },
  //                 data: {
  //                   columns: [{ name: "Skip", data_type_name: "INT8" }],
  //                   items: [{ Skip: summary.skip }],
  //                 },
  //               },
  //               {
  //                 name: `${props.name}.counter.info`,
  //                 node_type: "counter",
  //                 properties: { style: "info" },
  //                 data: {
  //                   columns: [{ name: "Info", data_type_name: "INT8" }],
  //                   items: [{ Info: summary.info }],
  //                 },
  //               },
  //               {
  //                 name: `${props.name}.counter.alarm`,
  //                 node_type: "counter",
  //                 properties: { style: "alert" },
  //                 data: {
  //                   columns: [{ name: "Alarm", data_type_name: "INT8" }],
  //                   items: [{ Alarm: summary.alarm }],
  //                 },
  //               },
  //               {
  //                 name: `${props.name}.counter.error`,
  //                 node_type: "counter",
  //                 properties: { style: "alert" },
  //                 data: {
  //                   columns: [{ name: "Error", data_type_name: "INT8" }],
  //                   items: [{ Error: summary.error }],
  //                 },
  //               },
  //             ],
  //           },
  //         ],
  //       },
  //     ]}
  //   />
  // );
};

export default Control;
