import Primitive from "../../Primitive";
import { isArray } from "lodash";

const findColumnIndices = (data) => {
  const headerRow = data[0];
  return {
    status: headerRow.findIndex((header) => header.toLowerCase() === "status"),
    total: headerRow.findIndex((header) => header.toLowerCase() === "total"),
  };
};

const prepareData = (data) => {
  if (!data || !isArray(data) || data.length === 0) {
    return null;
  }
  const columnIndices = findColumnIndices(data);
  const okRow = data
    .slice(1)
    .find((d) => d[columnIndices.status].toLowerCase() === "ok");
  const alarmRow = data
    .slice(1)
    .find((d) => d[columnIndices.status].toLowerCase() === "alarm");

  if (!alarmRow && !okRow) {
    return null;
  }

  const alarm = alarmRow ? parseInt(alarmRow[columnIndices.total]) : 0;
  const ok = okRow ? parseInt(okRow[columnIndices.total]) : 0;

  return {
    columnIndices,
    alarm,
    ok,
    total: alarm + ok,
  };
};

const ControlProgress = ({ data, error, title }) => {
  const preparedData = prepareData(data);
  return (
    <Primitive error={error} ready={!!preparedData}>
      {/*<div className="flex w-full">{preparedData.ok}</div>*/}
      <div className="flex w-full">
        {preparedData.ok && (
          <div
            className="rounded-l-md bg-ok text-white text-center p-1"
            style={{
              width: `${(preparedData.ok / preparedData.total) * 100}%`,
            }}
          >
            {preparedData.ok} compliant buckets
            {/*{(preparedData.ok / preparedData.total) * 100}%*/}
          </div>
        )}
        {preparedData.alarm && (
          <div
            className="rounded-r-md bg-alert text-white text-center p-1"
            style={{
              width: `${(preparedData.alarm / preparedData.total) * 100}%`,
            }}
          >
            {preparedData.alarm > 0 ? preparedData.alarm : null} non-compliant
            buckets
            {/*{(preparedData.alarm / preparedData.total) * 100}%*/}
          </div>
        )}
      </div>
    </Primitive>
  );
};

export default {
  type: "control_progress",
  component: ControlProgress,
};
