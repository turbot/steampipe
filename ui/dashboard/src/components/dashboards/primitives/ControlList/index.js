import Primitive from "../../Primitive";

const getField = (data, rowId, colName, colIndex, defaultValue) => {
  const headers = {};
  data[0].forEach((v, k) => {
    headers[k] = k;
    headers[v] = k;
  });
  var lookup = headers[colName] ?? headers[colIndex];
  var result = data[rowId][lookup];
  if (result === undefined) {
    result = defaultValue;
  }
  return result;
};

const statusColor = (status) => {
  switch (status) {
    case "OK":
      return "bg-green-600";
    case "ALARM":
      return "bg-red-600";
    default:
      return "bg-gray-600";
  }
};

const ControlList = ({ definition }) => {
  const { data, error } = definition;
  const rows = [];

  if (!error && data) {
    for (var rowId = 1; rowId < data.length; rowId++) {
      var status = getField(data, rowId, "status", 0, "TBD");
      var title = getField(data, rowId, "title", 1, "Unknown");
      var reason = getField(data, rowId, "reason", 2, "");
      rows.push(
        <li
          key={rowId}
          className="relative col-span-1 flex shadow-sm rounded-md"
        >
          <div
            className={
              statusColor(status) +
              " flex-shrink-0 flex items-center justify-center w-24 text-white text-sm font-medium rounded-l-md"
            }
          >
            {status}
          </div>
          <div className="flex-1 flex pl-2 items-center justify-between border-t border-r border-b border-gray-200 bg-white rounded-r-md truncate">
            <div className="px-2 py-2 text-sm truncate">
              <a href="#" className="text-gray-900 hover:text-gray-600">
                {title}
              </a>
            </div>
            {reason && (
              <div className="flex-grow px-2 py-2 text-sm truncate">
                <i className="text-gray-500">{reason}</i>
              </div>
            )}
          </div>
        </li>
      );
    }
  }

  if (rows.length === 0) {
    rows.push(
      <li key={0} className="relative col-span-1 flex shadow-sm rounded-md">
        <div className="flex-1 px-4 py-2 bg-white text-sm text-gray-400 border italic rounded-md">
          {(definition.option && definition.options.empty_text) ||
            "No results."}
        </div>
      </li>
    );
  }

  return (
    <Primitive error={error} ready={!!data}>
      <ul className="grid gap-1">{rows}</ul>
    </Primitive>
  );
};

export default {
  type: "control_list",
  component: ControlList,
};

/*

<>

<ul className="grid gap-1 mt-3">

  <li className="relative col-span-1 flex shadow-sm rounded-md">
    <div className="flex-shrink-0 flex items-center justify-center w-24 bg-red-600 text-white text-sm font-medium rounded-l-md">
      ALARM
    </div>
    <div className="flex-1 flex items-center justify-between border-t border-r border-b border-gray-200 bg-white rounded-r-md truncate">
      <div className="flex-1 px-4 py-2 text-sm truncate">
        <a href="#" className="text-gray-900 hover:text-gray-600">
          my-data-bucket
        </a>
      </div>
    </div>
  </li>

</ul>

<div className="flex flex-col">
  <div className="overflow-x-auto sm:-mx-6 lg:-mx-8">
    <div className="py-2 align-middle inline-block min-w-full sm:px-6 lg:px-8">
      <div className="shadow overflow-hidden border-b border-gray-200 sm:rounded-lg">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              {data[0].map((col, index) => (
                <th
                  scope="col"
                  key={index}
                  className="px-6 py-3 text-left text-sm font-medium text-gray-500 tracking-wider"
                >
                  {col}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
              {data.slice(1).map((row, rowIndex) => (
                <tr key={rowIndex}>
                  <td key="a" className="hidden px-6 py-4 whitespace-nowrap text-sm text-gray-500 md:block">
                    {rowIndex % 2 === 0 ?


    <div className="flex-shrink-0 bg-indigo-500 rounded-md p-3 mr-5 hidden">
      <svg className="h-6 w-6 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
      </svg>

/*
                    <span className="inline-flex items-center px-2 py-0.5 rounded-sm text-xs font-medium bg-green-600 text-white capitalize">

<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
</svg>

                      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
</svg>

                    </span>
*/

/*
    </div>

                    :
                    <span className="inline-flex items-center w-16 py-0.5 rounded-sm text-xs font-medium bg-red-600 text-white capitalize">
                      ALARM
                    </span>
                    }
                  </td>
                  {row.map((cell, cellIndex) => (
                    <td key={cellIndex} className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {cell}
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

</>

      )}
    </Primitive>
  );
};

*/

/*
    <div className="flex-shrink-0 pr-2">
      <button id="pinned-project-options-menu-0" aria-haspopup="true" className="w-8 h-8 bg-white inline-flex items-center justify-center text-gray-400 rounded-full hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-purple-500">
        <span className="sr-only">Open options</span>
        <svg className="w-5 h-5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
          <path d="M10 6a2 2 0 110-4 2 2 0 010 4zM10 12a2 2 0 110-4 2 2 0 010 4zM10 18a2 2 0 110-4 2 2 0 010 4z" />
        </svg>
      </button>
      <div className="z-10 mx-3 origin-top-right absolute right-10 top-3 w-48 mt-1 rounded-md shadow-lg bg-white ring-1 ring-black ring-opacity-5 divide-y divide-gray-200" role="menu" aria-orientation="vertical" aria-labelledby="pinned-project-options-menu-0">
        <div className="py-1" role="none">
          <a href="#" className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 hover:text-gray-900" role="menuitem">View</a>
        </div>
        <div className="py-1" role="none">
          <a href="#" className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 hover:text-gray-900" role="menuitem">Removed from pinned</a>
          <a href="#" className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 hover:text-gray-900" role="menuitem">Share</a>
        </div>
      </div>
    </div>
*/

/*
<div className="flex flex-col">
  <div className="overflow-x-auto sm:-mx-6 lg:-mx-8">
    <div className="py-2 align-middle inline-block min-w-full sm:px-6 lg:px-8">
      <div className="shadow overflow-hidden border-b border-gray-200 sm:rounded-lg">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              {data[0].map((col, index) => (
                <th
                  scope="col"
                  key={index}
                  className="px-6 py-3 text-left text-sm font-medium text-gray-500 tracking-wider"
                >
                  {col}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
              {data.slice(1).map((row, rowIndex) => (
                <tr key={rowIndex}>
                  <td key="a" className="hidden px-6 py-4 whitespace-nowrap text-sm text-gray-500 md:block">
                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800 capitalize">
                      OK
                    </span>
                  </td>
                  {row.map((cell, cellIndex) => (
                    <td key={cellIndex} className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {cell}
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

/*
        <table className="table-auto w-full text-left">
          <thead>
            <tr>
              {data[0].map((col, index) => (
                <th
                  key={index}
                  className="border px-4 py-2 font-semibold"
                >
                  {col}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="text-gray-600">
            {data.slice(1).map((row, rowIndex) => (
              <tr key={rowIndex}>
                {row.map((cell, cellIndex) => (
                  <td key={cellIndex} className="border px-4 py-2 align-text-top">
                    {cell}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
        */
