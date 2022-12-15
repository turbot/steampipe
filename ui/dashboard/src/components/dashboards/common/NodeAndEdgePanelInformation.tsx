import ErrorMessage from "../../ErrorMessage";
import Icon from "../../Icon";
import LoadingIndicator from "../LoadingIndicator";
import { CategoryStatus } from "../graphs/types";
import { DashboardRunState } from "../../../types";
import { Fragment } from "react";
import { Node } from "reactflow";

type NodeAndEdgePanelInformationProps = {
  nodes: Node[];
  status: DashboardRunState;
  pendingWiths: CategoryStatus[];
  errorWiths: CategoryStatus[];
  completeWiths: CategoryStatus[];
  pendingCategories: CategoryStatus[];
  errorCategories: CategoryStatus[];
  completeCategories: CategoryStatus[];
};

const NodeAndEdgePanelInformation = ({
  nodes,
  status,
  pendingWiths,
  errorWiths,
  completeWiths,
  pendingCategories,
  errorCategories,
  completeCategories,
}: NodeAndEdgePanelInformationProps) => {
  return (
    <div className="space-y-2 overflow-y-scroll">
      <div className="space-y-1">
        <span className="block font-medium">Withs</span>
        <div>
          {completeWiths.length} complete, {pendingWiths.length} running,{" "}
          {errorWiths.length} {errorWiths.length === 1 ? "error" : "errors"}
        </div>
        {pendingWiths.length === 0 &&
          errorWiths.length === 0 &&
          status === "complete" &&
          nodes.length === 0 && (
            <span className="block text-foreground-light italic">
              No nodes or edges
            </span>
          )}
        {pendingWiths.map((withStatus) => (
          <div key={withStatus.id} className="flex items-center space-x-1">
            <LoadingIndicator className="w-3.5 h-3.5" />
            <span key={withStatus.id} className="block">
              {withStatus.title || withStatus.id}
            </span>
          </div>
        ))}
        {errorWiths.map((withStatus) => (
          <Fragment key={withStatus.id}>
            <div className="flex items-center space-x-1">
              <Icon
                className="w-3.5 h-3.5 text-alert"
                icon="materialsymbols-solid:error"
              />
              <span key={withStatus.id} className="block">
                {withStatus.title || withStatus.id}
              </span>
            </div>
            <span className="block">
              <ErrorMessage error={withStatus.error} />{" "}
            </span>
          </Fragment>
        ))}
      </div>
      <div className="space-y-1">
        <span className="block font-medium">Categories</span>
        <div>
          {completeCategories.length} complete, {pendingCategories.length}{" "}
          running, {errorCategories.length}{" "}
          {errorCategories.length === 1 ? "error" : "errors"}
        </div>
        {pendingCategories.length === 0 &&
          errorCategories.length === 0 &&
          status === "complete" &&
          nodes.length === 0 && (
            <span className="block text-foreground-light italic">
              No nodes or edges
            </span>
          )}
        {pendingCategories.map((category) => (
          <div key={category.id} className="flex items-center space-x-1">
            <LoadingIndicator className="w-3.5 h-3.5" />
            <span key={category.id} className="block">
              {category.title || category.id}
            </span>
          </div>
        ))}
        {errorCategories.map((category) => (
          <Fragment key={category.id}>
            <div className="flex items-center space-x-1">
              <Icon
                className="w-3.5 h-3.5 text-alert"
                icon="materialsymbols-solid:error"
              />
              <span key={category.id} className="block">
                {category.title || category.id}
              </span>
            </div>
            {category.nodesInError?.map((n) => (
              <span key={n.id} className="block">
                <ErrorMessage error={n.error} />{" "}
              </span>
            ))}
            {category.edgesInError?.map((e) => (
              <span key={e.id} className="block">
                <ErrorMessage error={e.error} />{" "}
              </span>
            ))}
          </Fragment>
        ))}
      </div>
    </div>
  );
};

export default NodeAndEdgePanelInformation;
