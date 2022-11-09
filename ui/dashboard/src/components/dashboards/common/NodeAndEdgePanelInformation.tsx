import Icon from "../../Icon";
import LoadingIndicator from "../LoadingIndicator";
import { NodeStatus } from "../graphs/types";

type NodeAndEdgePanelInformationProps = {
  pendingNodes: NodeStatus[];
  errorNodes: NodeStatus[];
  completeNodes: NodeStatus[];
};

const NodeAndEdgePanelInformation = ({
  pendingNodes,
  errorNodes,
  completeNodes,
}: NodeAndEdgePanelInformationProps) => {
  return (
    <div className="space-y-2 overflow-y-scroll">
      <div className="space-y-1">
        {(pendingNodes.length > 0 || errorNodes.length > 0) && (
          <span className="block font-medium">Nodes</span>
        )}
        {pendingNodes.map((n) => (
          <div className="flex items-center space-x-1">
            <LoadingIndicator className="w-3 h-3" />
            <span key={n.id} className="block">
              {n.title}
            </span>
          </div>
        ))}
        {errorNodes.map((n) => (
          <div className="flex items-center space-x-1">
            <LoadingIndicator className="w-3 h-3" />
            <span key={n.id} className="block">
              {n.title}
            </span>
          </div>
        ))}
        {completeNodes.map((n) => (
          <div className="flex items-center space-x-1">
            <Icon className="w-3 h-3 text-ok" icon="check" />
            <span key={n.id} className="block">
              {n.title}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
};

export default NodeAndEdgePanelInformation;
