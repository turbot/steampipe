import ErrorMessage from "../../ErrorMessage";
import Icon from "../../Icon";
import LoadingIndicator from "../LoadingIndicator";
import { CategoryStatus } from "../graphs/types";
import { Fragment } from "react";

type NodeAndEdgePanelInformationProps = {
  pendingCategories: CategoryStatus[];
  errorCategories: CategoryStatus[];
  completeCategories: CategoryStatus[];
};

const NodeAndEdgePanelInformation = ({
  pendingCategories,
  errorCategories,
  completeCategories,
}: NodeAndEdgePanelInformationProps) => {
  return (
    <div className="space-y-2 overflow-y-scroll">
      <div className="space-y-1">
        {(pendingCategories.length > 0 || errorCategories.length > 0) && (
          <span className="block font-medium">Categories</span>
        )}
        <div>
          {completeCategories.length} complete, {pendingCategories.length}{" "}
          running, {errorCategories.length}{" "}
          {errorCategories.length === 1 ? "error" : "errors"}
        </div>
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
                icon="heroicons-solid:exclamation-circle"
              />
              <span key={category.id} className="block">
                {category.title || category.id}
              </span>
            </div>
            {category.nodesInError?.map((n) => (
              <span key={n.id}>
                <ErrorMessage error={n.error} />{" "}
              </span>
            ))}
            {category.edgesInError?.map((e) => (
              <span key={e.id}>
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
