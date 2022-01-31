import CheckCounter from "../common/CheckCounter";
import LayoutPanel from "../../layout/common/LayoutPanel";
import { CheckLeafNodeDataGroupSummaryStatus, CheckProps } from "../common";
import { useMemo } from "react";

const Benchmark = (props: CheckProps) => {
  const { loading, summary } = useMemo(() => {
    const summary = props.execution_tree?.root?.summary?.status;
    if (!summary) {
      return {
        loading: true,
        summary: {} as CheckLeafNodeDataGroupSummaryStatus,
      };
    }
    return { loading: false, summary };
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
    </LayoutPanel>
  );
};

export default Benchmark;
