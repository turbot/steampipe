import CheckPanel from "../CheckPanel";
import { CheckNode, CheckSummary } from "../common";

interface CheckGroupingProps {
  node: CheckNode;
  rootSummary: CheckSummary;
}

const CheckGrouping = ({ node, rootSummary }: CheckGroupingProps) => {
  return (
    <div className="space-y-4 md:space-y-6 col-span-12">
      {node.children?.map((child) => (
        <CheckPanel key={child.name} node={child} rootSummary={rootSummary} />
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
