import Icon from "../../../Icon";
import Primitive from "../../Primitive";
import { alarmIcon, okIcon, tbdIcon } from "../../../../constants/icons";
import { isObject } from "lodash";

const renderCell = (cell) => {
  if (isObject(cell)) {
    return JSON.stringify(cell, null, 2);
  }
  return cell;
};

const ControlTable = ({ definition }) => {
  const { data, error } = definition;

  return (
    <Primitive error={error} ready={!!data}>
      {data && (
        <table className="min-w-full divide-y divide-table-divide border-t border-table-border overflow-hidden">
          <thead className="bg-table-head text-table-head">
            <tr>
              {data.length > 0 &&
                data[0].map((col, index) => (
                  <th
                    scope="col"
                    key={index}
                    className="px-4 sm:px-6 py-3 text-left text-xs font-normal tracking-wider"
                  >
                    {col}
                  </th>
                ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-table-divide">
            {data.length === 1 && (
              <tr>
                <td
                  className="px-4 sm:px-6 py-4 align-top content-center text-sm italic"
                  colSpan={data[0].length}
                >
                  No results
                </td>
              </tr>
            )}
            {data.length > 1 &&
              data.slice(1).map((row, rowIndex) => (
                <tr key={rowIndex}>
                  {row.map((cell, cellIndex) => {
                    const cellTitle = data[0][cellIndex] || "";
                    return (
                      <td
                        key={cellIndex}
                        className="px-4 sm:px-6 py-4 align-top content-center text-sm text-foreground"
                      >
                        {cellTitle.toLowerCase() === "status" && (
                          <>
                            {["ALARM", "ERROR", "INVALID"].includes(
                              cell.toUpperCase()
                            ) && (
                              <div className="whitespace-nowrap text-alert">
                                <Icon icon={alarmIcon} />
                              </div>
                            )}
                            {cell.toUpperCase() === "OK" && (
                              <div className="whitespace-nowrap text-ok">
                                <Icon icon={okIcon} />
                              </div>
                            )}
                            {cell.toUpperCase() === "TBD" && (
                              <div className="whitespace-nowrap text-tbd">
                                <Icon icon={tbdIcon} />
                              </div>
                            )}
                          </>
                        )}

                        {cellTitle.toLowerCase() === "reason" && (
                          <div className="text-black-scale-6 italic">
                            {cell}
                          </div>
                        )}

                        {cellTitle.toLowerCase() !== "status" &&
                          cellTitle.toLowerCase() !== "reason" &&
                          renderCell(cell)}
                      </td>
                    );
                  })}
                </tr>
              ))}
          </tbody>
        </table>
      )}
    </Primitive>
  );
};

export default {
  type: "control_table",
  component: ControlTable,
};

/*

<div className="flex flex-col">
  <div className="overflow-x-auto sm:-mx-6 lg:-mx-8">
    <div className="py-2 align-middle inline-block min-w-full sm:px-6 lg:px-8">
      <div className="shadow border-t border-gray-200 sm:rounded-lg overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50 text-gray-500">
            <tr>
              {data[0].map((col, index) => (
                <th
                  scope="col"
                  key={index}
                  className="px-4 sm:px-6 py-3 text-left text-xs font-normal tracking-wider"
                >
                  {col}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
              {data.slice(1).map((row, rowIndex) => (
                <tr key={rowIndex}>
                  {row.map((cell, cellIndex) => (
                    <td key={cellIndex} className="px-4 sm:px-6 py-4 align-top content-center text-sm text-gray-900">

                      {["ALARM", "ERROR", "INVALID"].includes(cell) &&
                        <div className="text-center whitespace-nowrap">
                          {rowIndex % 7 === 0 ?
                            <span className="hidden text-yellow-900 mr-2">&#9888;</span>
                            :
                            <span className="hidden text-white mr-2">&#9888;</span>
                          }
                          <span className="h-4 w-4 bg-red-100 rounded-full inline-flex items-center justify-center" aria-hidden="true">
                            <span className="h-2.5 w-2.5 bg-red-600 rounded-full"></span>
                          </span>
                        </div>
                      }

                      {[].includes(cell) &&
                        <div className="text-center whitespace-nowrap">
                          <span className="text-red-600">
                            <span className="font-bold">&#215;</span>
                            <span className="hidden ml-2 text-gray-300 text-xs">{cell}</span>
                          </span>
                        </div>
                      }

                      {[].includes(cell) &&
                        <div className="text-center whitespace-nowrap align-middle">
                          <span className="bg-red-600 text-white -my-1 py-1 px-2 rounded-md">
                            <span className="font-bold">&#215;</span>
                            <span className="ml-2 font-semibold">{cell}</span>
                          </span>
                        </div>
                      }

                      {cell === "OK" &&
                        <div className="text-center whitespace-nowrap">
                          <span className="text-green-600">
                            <span className="font-bold">&#10003;</span>
                            <span className="hidden ml-2 text-gray-300 text-xs">{cell}</span>
                          </span>
                        </div>
                      }

                      {cellIndex === 3 &&
                        <div className="text-gray-400 italic">{cell}</div>
                      }

                      {!["ALARM", "INVALID", "ERROR", "OK"].includes(cell) && cellIndex !== 3 &&
                        <>
                        {cellIndex === 2 && rowIndex % 7 === 0 &&
                          <span className="text-yellow-900 bg-yellow-100 rounded-sm text-xs px-2 py-1 mr-2">HIGH</span>
                        }
                        {cell}
                        </>
                      }
                    </td>
                  ))}
                </tr>
              ))}
          </tbody>
        </table>
      </div>
    </div>
  </div>
</div>

*/
