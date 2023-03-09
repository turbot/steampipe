import ErrorMessage from "../../ErrorMessage";
import Icon from "../../Icon";
import LoadingIndicator from "../LoadingIndicator";
import { DashboardRunState } from "../../../types";
import { EdgeStatus, GraphStatuses, NodeStatus } from "../graphs/types";
import { Node } from "reactflow";

type NodeAndEdgePanelInformationProps = {
  nodes: Node[];
  status: DashboardRunState;
  statuses: GraphStatuses;
};

const nodeOrEdgeTitle = (nodeOrEdge: NodeStatus | EdgeStatus) =>
  nodeOrEdge.title ||
  nodeOrEdge?.category?.title ||
  nodeOrEdge?.category?.name ||
  nodeOrEdge.id;

const WaitingRow = ({ title }) => (
  <div className="flex items-center space-x-1">
    <Icon
      className="w-3.5 h-3.5 text-foreground-light shrink-0"
      icon="pending"
    />
    <span className="block truncate">{title}</span>
  </div>
);

const RunningRow = ({ title }) => (
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
  statuses,
}: NodeAndEdgePanelInformationProps) => (
  <div className="space-y-2 overflow-y-scroll">
    <div className="space-y-1">
      <div>
        {statuses.complete.total} complete, {statuses.running.total} running,{" "}
        {statuses.blocked.total} waiting, {statuses.error.total}{" "}
        {statuses.error.total === 1 ? "error" : "errors"}.
      </div>
      {statuses.initialized.total === 0 &&
        statuses.blocked.total === 0 &&
        statuses.running.total === 0 &&
        statuses.complete.total === 0 &&
        status === "complete" &&
        nodes.length === 0 && (
          <span className="block text-foreground-light italic">
            No nodes or edges
          </span>
        )}
      {statuses.running.withs.map((withStatus, idx) => (
        <RunningRow
          key={`with:${withStatus.id}-${idx}`}
          title={`with: ${withStatus.title || withStatus.id}`}
        />
      ))}
      {statuses.running.nodes.map((node, idx) => (
        <RunningRow
          key={`node:${node.id}-${idx}`}
          title={`node: ${nodeOrEdgeTitle(node)}`}
        />
      ))}
      {statuses.running.edges.map((edge, idx) => (
        <RunningRow
          key={`edge:${edge.id}-${idx}`}
          title={`edge: ${nodeOrEdgeTitle(edge)}`}
        />
      ))}
      {statuses.blocked.withs.map((withStatus, idx) => (
        <WaitingRow
          key={`with:${withStatus.id}-${idx}`}
          title={`with: ${withStatus.title || withStatus.id}`}
        />
      ))}
      {statuses.blocked.nodes.map((node, idx) => (
        <WaitingRow
          key={`node:${node.id}-${idx}`}
          title={`node: ${nodeOrEdgeTitle(node)}`}
        />
      ))}
      {statuses.blocked.edges.map((edge, idx) => (
        <WaitingRow
          key={`edge:${edge.id}-${idx}`}
          title={`edge: ${nodeOrEdgeTitle(edge)}`}
        />
      ))}
      {statuses.error.withs.map((withStatus, idx) => (
        <ErrorRow
          key={`with:${withStatus.id}-${idx}`}
          title={`with: ${withStatus.title || withStatus.id}`}
          error={withStatus.error}
        />
      ))}
      {statuses.error.nodes.map((node, idx) => (
        <ErrorRow
          key={`node:${node.id}-${idx}`}
          title={`node: ${nodeOrEdgeTitle(node)}`}
          error={node.error}
        />
      ))}
      {statuses.error.edges.map((edge, idx) => (
        <ErrorRow
          key={`edge:${edge.id}-${idx}`}
          title={`edge: ${nodeOrEdgeTitle(edge)}`}
          error={edge.error}
        />
      ))}
    </div>
  </div>
);

export default NodeAndEdgePanelInformation;
