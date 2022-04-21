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
        <div className="flex justify-between">
          <div className="flex justify-between">
            <h3
              id={`${node.name}-title`}
              className="truncate"
              title={node.title}
            >
              {node.title}
            </h3>
          </div>
        </div>
      </section>
    </div>
  );
};

export default CheckPanel;
