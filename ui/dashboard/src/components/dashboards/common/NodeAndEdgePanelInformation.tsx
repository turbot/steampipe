import ErrorMessage from "../../ErrorMessage";
import Icon from "../../Icon";
import LoadingIndicator from "../LoadingIndicator";
import { DashboardRunState } from "../../../types";
import { EdgeStatus, NodeStatus, WithStatus } from "../graphs/types";
import { Node } from "reactflow";

type NodeAndEdgePanelInformationProps = {
  nodes: Node[];
  status: DashboardRunState;
  pendingWiths: WithStatus[];
  pendingNodes: NodeStatus[];
  pendingEdges: EdgeStatus[];
  errorWiths: WithStatus[];
  errorNodes: NodeStatus[];
  errorEdges: EdgeStatus[];
  completeWiths: WithStatus[];
  completeNodes: NodeStatus[];
  completeEdges: EdgeStatus[];
};

const nodeOrEdgeTitle = (nodeOrEdge: NodeStatus | EdgeStatus) =>
  nodeOrEdge.title ||
  nodeOrEdge?.category?.title ||
  nodeOrEdge?.category?.name ||
  nodeOrEdge.id;

const PendingRow = ({ title }) => (
  <div className="flex items-center space-x-1">
    <LoadingIndicator className="w-3.5 h-3.5 shrink-0" />
    <span className="block truncate">{title}</span>
  </div>
);

const ErrorRow = ({ title, error }: { title: string; error?: string }) => (
  <>
    <div className="flex items-center space-x-1">
      <Icon
        className="w-3.5 h-3.5 text-alert shrink-0"
        icon="materialsymbols-solid:error"
      />
      <span className="block">{title}</span>
    </div>
    {error && (
      <span className="block">
        <ErrorMessage error={error} />
      </span>
    )}
  </>
);

const NodeAndEdgePanelInformation = ({
  nodes,
  status,
  pendingWiths,
  pendingNodes,
  pendingEdges,
  errorWiths,
  errorNodes,
  errorEdges,
  completeWiths,
  completeNodes,
  completeEdges,
}: NodeAndEdgePanelInformationProps) => {
  const totalPending =
    pendingWiths.length + pendingNodes.length + pendingEdges.length;
  const totalError = errorWiths.length + errorNodes.length + errorEdges.length;
  const totalComplete =
    completeWiths.length + completeNodes.length + completeEdges.length;
  return (
    <div className="space-y-2 overflow-y-scroll">
      <div className="space-y-1">
        <div>
          {totalComplete} complete, {totalPending} running, {totalError}{" "}
          {totalError === 1 ? "error" : "errors"}
        </div>
        {totalPending === 0 &&
          totalError === 0 &&
          status === "complete" &&
          nodes.length === 0 && (
            <span className="block text-foreground-light italic">
              No nodes or edges
            </span>
          )}
        {pendingWiths.map((withStatus) => (
          <PendingRow
            key={withStatus.id}
            title={`with: ${withStatus.title || withStatus.id}`}
          />
        ))}
        {pendingNodes.map((node) => (
          <PendingRow key={node.id} title={`node: ${nodeOrEdgeTitle(node)}`} />
        ))}
        {pendingEdges.map((edge) => (
          <PendingRow key={edge.id} title={`edge: ${nodeOrEdgeTitle(edge)}`} />
        ))}
        {errorWiths.map((withStatus) => (
          <ErrorRow
            key={withStatus.id}
            title={`with: ${withStatus.title || withStatus.id}`}
            error={withStatus.error}
          />
        ))}
        {errorNodes.map((node) => (
          <ErrorRow
            key={node.id}
            title={`node: ${nodeOrEdgeTitle(node)}`}
            error={node.error}
          />
        ))}
        {errorEdges.map((edge) => (
          <ErrorRow
            key={edge.id}
            title={`edge: ${nodeOrEdgeTitle(edge)}`}
            error={edge.error}
          />
        ))}
      </div>
    </div>
  );
};

export default NodeAndEdgePanelInformation;
