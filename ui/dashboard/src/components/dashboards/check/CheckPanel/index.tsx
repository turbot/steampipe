import CheckSummaryChart from "../CheckSummaryChart";
import { CheckNode } from "../common";
import { useState } from "react";
import { ArrowDownIcon, ChevronDownIcon } from "@heroicons/react/solid";

interface CheckPanelProps {
  node: CheckNode;
}

const CheckPanel = ({ node }: CheckPanelProps) => {
  const [expanded, setExpanded] = useState(false);

  return (
    <div id={node.name}>
      <section className="bg-dashboard-panel shadow-sm rounded-md p-4">
        <div className="flex justify-between items-center space-x-4">
          <div className="flex flex-grow justify-between items-center">
            <h3
              id={`${node.name}-title`}
              className="truncate"
              title={node.title}
            >
              {node.title}
            </h3>
            <CheckSummaryChart name={node.name} summary={node.summary} />
          </div>
          <ChevronDownIcon className="h-5 w-5 flex-shrink-0 text-foreground-lightest" />
        </div>
      </section>
    </div>
  );
};

export default CheckPanel;
