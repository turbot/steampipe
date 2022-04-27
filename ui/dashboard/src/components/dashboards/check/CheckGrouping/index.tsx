import CheckPanel from "../CheckPanel";
import sortBy from "lodash/sortBy";
import { CheckDisplayGroup, CheckNode, CheckSummary } from "../common";

interface CheckGroupingProps {
  node: CheckNode;
  groupingConfig: CheckDisplayGroup[];
  firstChildSummaries: CheckSummary[];
}

const CheckGrouping = ({
  node,
  groupingConfig,
  firstChildSummaries,
}: CheckGroupingProps) => (
  <div className="space-y-4 md:space-y-6 col-span-12">
    {sortBy(node.children, "sort")?.map((child) => (
      <CheckPanel
        key={child.name}
        depth={1}
        node={child}
        groupingConfig={groupingConfig}
        firstChildSummaries={firstChildSummaries}
      />
    ))}
  </div>
);

export default CheckGrouping;

// TODO
// Summary chart should show something if no results
// Add counts to summary chart
// Add animation to summary charts + remove from row
// If no results and no error, show no resources row
// Scaling of summary charts is off

// "benchmark" / "control" / "result"
// <dimension> / "benchmark" / "control" / "result"
// <tag> / "benchmark" / "control" / "result"
// etc
