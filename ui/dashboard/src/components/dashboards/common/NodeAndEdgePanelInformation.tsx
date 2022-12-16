import ErrorMessage from "../../ErrorMessage";
import Icon from "../../Icon";
import LoadingIndicator from "../LoadingIndicator";
import {
  CategoryStatus,
  EdgeStatus,
  NoCategoryStatus,
  NodeStatus,
  WithStatus,
} from "../graphs/types";
import { DashboardRunState } from "../../../types";
import { Node } from "reactflow";

type NodeAndEdgePanelInformationProps = {
  nodes: Node[];
  status: DashboardRunState;
  pendingWiths: WithStatus[];
  pendingNoCategories: NoCategoryStatus[];
  pendingCategories: CategoryStatus[];
  errorWiths: WithStatus[];
  errorNoCategories: NoCategoryStatus[];
  errorCategories: CategoryStatus[];
  completeWiths: WithStatus[];
  completeNoCategories: NoCategoryStatus[];
  completeCategories: CategoryStatus[];
};

const PendingRow = ({ title }) => (
  <div className="flex items-center space-x-1">
    <LoadingIndicator className="w-3.5 h-3.5" />
    <span className="block">{title}</span>
  </div>
);

const ErrorRow = ({
  title,
  error,
  nodesInError,
  edgesInError,
}: {
  title: string;
  error?: string;
  nodesInError?: NodeStatus[];
  edgesInError?: EdgeStatus[];
}) => (
  <>
    <div className="flex items-center space-x-1">
      <Icon
        className="w-3.5 h-3.5 text-alert"
        icon="materialsymbols-solid:error"
      />
      <span className="block">{title}</span>
    </div>
    {error && (
      <span className="block">
        <ErrorMessage error={error} />
      </span>
    )}
    {nodesInError?.map((n) => (
      <span key={n.id} className="block">
        <ErrorMessage error={n.error} />{" "}
      </span>
    ))}
    {edgesInError?.map((e) => (
      <span key={e.id} className="block">
        <ErrorMessage error={e.error} />{" "}
      </span>
    ))}
  </>
);

const NodeAndEdgePanelInformation = ({
  nodes,
  status,
  pendingWiths,
  pendingNoCategories,
  pendingCategories,
  errorWiths,
  errorNoCategories,
  errorCategories,
  completeWiths,
  completeNoCategories,
  completeCategories,
}: NodeAndEdgePanelInformationProps) => {
  const totalPending =
    pendingWiths.length + pendingNoCategories.length + pendingCategories.length;
  const totalError =
    errorWiths.length + errorNoCategories.length + errorCategories.length;
  const totalComplete =
    completeWiths.length +
    completeNoCategories.length +
    completeCategories.length;
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
        {pendingNoCategories.map((noCategory) => (
          <PendingRow
            key={noCategory.id}
            title={`${noCategory.panelType}: ${
              noCategory.title || noCategory.id
            }`}
          />
        ))}
        {pendingCategories.map((category) => (
          <PendingRow
            key={category.id}
            title={`category: ${category.title || category.id}`}
          />
        ))}
        {errorWiths.map((withStatus) => (
          <ErrorRow
            key={withStatus.id}
            title={`with: ${withStatus.title || withStatus.id}`}
            error={withStatus.error}
          />
        ))}
        {errorNoCategories.map((noCategory) => (
          <ErrorRow
            key={noCategory.id}
            title={`${noCategory.panelType}: ${
              noCategory.title || noCategory.id
            }`}
            error={noCategory.error}
          />
        ))}
        {errorCategories.map((category) => (
          <ErrorRow
            key={category.id}
            title={`category: ${category.title || category.id}`}
            nodesInError={category.nodesInError}
            edgesInError={category.edgesInError}
          />
        ))}
      </div>
    </div>
  );
};

export default NodeAndEdgePanelInformation;
