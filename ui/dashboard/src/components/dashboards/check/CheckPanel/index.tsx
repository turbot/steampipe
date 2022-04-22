import CheckSummaryChart from "../CheckSummaryChart";
import { CheckNode } from "../common";
import {
  CollapseBenchmarkIcon,
  ExpandCheckNodeIcon,
} from "../../../../constants/icons";
import { useState } from "react";

interface CheckPanelProps {
  node: CheckNode;
}

const getMargin = (depth) => {
  switch (depth) {
    case 1:
      return "ml-[24px]";
    case 2:
      return "ml-[48px]";
    case 3:
      return "ml-[72px]";
    case 4:
      return "ml-[96px]";
    case 5:
      return "ml-[120px]";
    case 6:
      return "ml-[144px]";
    default:
      return "ml-0";
  }
};

const CheckChildren = ({ node }: CheckPanelProps) => {
  if (!node.children) {
    return null;
  }

  return (
    <>
      {node.children.map((child) => (
        <CheckPanel key={child.name} node={child} />
      ))}
    </>
  );
};

const CheckPanel = ({ node }: CheckPanelProps) => {
  const [expanded, setExpanded] = useState(false);

  return (
    <>
      <div id={node.name} className={getMargin(node.depth - 1)}>
        <section
          className="bg-dashboard-panel shadow-sm rounded-md p-4 cursor-pointer"
          onClick={() => setExpanded((current) => !current)}
        >
          <div className="flex items-center space-x-6">
            <div className="flex flex-grow justify-between items-center">
              <h3
                id={`${node.name}-title`}
                className="truncate mt-0"
                title={node.title}
              >
                {node.title}
              </h3>
              <CheckSummaryChart name={node.name} summary={node.summary} />
            </div>
            {!expanded && (
              <ExpandCheckNodeIcon className="h-7 w-7 flex-shrink-0 text-foreground-lightest" />
            )}
            {expanded && (
              <CollapseBenchmarkIcon className="h-7 w-7 flex-shrink-0 text-foreground-lightest" />
            )}
          </div>
        </section>
      </div>
      {expanded && <CheckChildren node={node} />}
    </>
  );
};

export default CheckPanel;
