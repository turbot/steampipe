import Primitive from "../../Primitive";
import { useMemo } from "react";

const ResourceDetail = ({ data, error }) => {
  const detail = useMemo(() => {
    if (error || !data) {
      return null;
    }

    if (data.length < 2) {
      return null;
    }

    const keys = data[0];
    const values = data[1];
    const detail = [];
    for (let i = 0; i < keys.length; i++) {
      const key = keys[i];
      const value = values[i];
      detail.push({ key, value });
    }
    return detail;
  }, [data, error]);
  return (
    <Primitive error={error} ready={!!data}>
      {detail &&
        detail.map((pair) => (
          <div key={pair.key} className="mb-2">
            <span className="block prose prose-sm font-light text-table-head">
              {pair.key}
            </span>
            <span className="block prose prose-sm">
              {pair.value !== null &&
                pair.value !== undefined &&
                pair.value.toString()}
              {(pair.value === null || pair.value === undefined) && (
                <span className="italic">No value</span>
              )}
            </span>
          </div>
        ))}
    </Primitive>
  );
};

export default {
  type: "resource_detail",
  component: ResourceDetail,
};
