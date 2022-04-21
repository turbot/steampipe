import CheckSummaryChart from "../CheckSummaryChart";
import { CheckNode } from "../common";
import { useState } from "react";

interface CheckPanelProps {
  node: CheckNode;
}

const CheckPanel = ({ node }: CheckPanelProps) => {
  const [expanded, setExpanded] = useState(false);

  return (
    <div id={node.name}>
      <section className="bg-dashboard-panel shadow-sm rounded-md p-4">
        <div className="flex justify-between items-center">
          <div className="flex justify-between items-center">
            <h3
              id={`${node.name}-title`}
              className="truncate"
              title={node.title}
            >
              {node.title}
            </h3>
          </div>
          <CheckSummaryChart name={node.name} summary={node.summary} />
        </div>
      </section>
    </div>
  );
};

export default CheckPanel;
