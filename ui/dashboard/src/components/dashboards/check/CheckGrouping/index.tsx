import CheckPanel from "../CheckPanel";
import sortBy from "lodash/sortBy";
import { CheckDisplayGroup, CheckNode, CheckSummary } from "../common";

interface CheckGroupingProps {
  node: CheckNode;
  groupingConfig: CheckDisplayGroup[];
  rootSummary: CheckSummary;
}

const CheckGrouping = ({
  node,
  groupingConfig,
  rootSummary,
}: CheckGroupingProps) => {
  return (
    <div className="space-y-4 md:space-y-6 col-span-12">
      {sortBy(node.children, "title")?.map((child) => (
        <CheckPanel
          key={child.name}
          node={child}
          groupingConfig={groupingConfig}
          rootSummary={rootSummary}
        />
      ))}
    </div>
  );
};

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
