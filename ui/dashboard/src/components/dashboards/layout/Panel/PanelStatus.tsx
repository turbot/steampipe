import ErrorMessage from "../../../ErrorMessage";
import Icon from "../../../Icon";
import LoadingIndicator from "../../LoadingIndicator";
import usePanelDependenciesStatus from "../../../../hooks/usePanelDependenciesStatus";
import { classNames } from "../../../../utils/styles";
import { HashLink } from "react-router-hash-link";
import { InputProperties } from "../../inputs/types";
import { PanelDefinition } from "../../../../types";
import { ReactNode } from "react";
import { useLocation } from "react-router-dom";

type PanelStatusProps = PanelStatusBaseProps & {
  definition: PanelDefinition;
  showPanelError: boolean;
};

type PanelStatusBaseProps = {
  definition: PanelDefinition;
};

type BasePanelStatusProps = {
  children: ReactNode;
  className?: string;
  definition: PanelDefinition;
};

const BasePanelStatus = ({
  children,
  className,
  definition,
}: BasePanelStatusProps) => (
  <div
    className={classNames(
      className,
      "w-full h-full p-4 break-keep rounded-md",
      !!definition.title ? "rounded-t-none" : null
    )}
  >
    {children}
  </div>
);

const PanelInitialized = ({ definition }: PanelStatusBaseProps) => {
  return (
    <BasePanelStatus definition={definition}>
      <div className="flex items-center space-x-1">
        <LoadingIndicator className="w-3.5 h-3.5 text-foreground-light shrink-0" />
        <span className="block truncate">Initialized...</span>
      </div>
    </BasePanelStatus>
  );
};

const PanelBlocked = ({ definition }) => {
  const panelDependenciesStatus = usePanelDependenciesStatus();
  const location = useLocation();
  return (
    <BasePanelStatus definition={definition}>
      <div className="flex items-center space-x-1">
        {panelDependenciesStatus.inputsAwaitingValue.length > 0 && (
          <>
            <Icon
              className="w-3.5 h-3.5 text-foreground-light shrink-0"
              icon="block"
            />
            <span className="block truncate">
              Awaiting input:{" "}
              <HashLink
                className="text-link"
                to={`${location.pathname}${
                  location.search ? location.search : ""
                }#${panelDependenciesStatus.inputsAwaitingValue[0].name}`}
              >
                {panelDependenciesStatus.inputsAwaitingValue[0].title ||
                  (
                    panelDependenciesStatus.inputsAwaitingValue[0]
                      .properties as InputProperties
                  ).unqualified_name}
              </HashLink>
            </span>
          </>
        )}
        {panelDependenciesStatus.inputsAwaitingValue.length === 0 &&
          panelDependenciesStatus.total === 0 && (
            <>
              <LoadingIndicator className="w-3.5 h-3.5 text-foreground-light shrink-0" />
              <span className="block truncate">Loading...</span>
            </>
          )}
        {panelDependenciesStatus.inputsAwaitingValue.length === 0 &&
          panelDependenciesStatus.total > 0 && (
            <>
              <LoadingIndicator className="w-3.5 h-3.5 text-foreground-light shrink-0" />
              <span className="block truncate">
                Loading{" "}
                {panelDependenciesStatus.status.complete.total +
                  panelDependenciesStatus.status.running.total +
                  1}{" "}
                of {panelDependenciesStatus.total + 1}...
              </span>
            </>
          )}
      </div>
    </BasePanelStatus>
  );
};

const PanelRunning = ({ definition }) => {
  const panelDependenciesStatus = usePanelDependenciesStatus();
  return (
    <BasePanelStatus definition={definition}>
      <div className="flex items-center space-x-1">
        {panelDependenciesStatus.inputsAwaitingValue.length === 0 &&
          panelDependenciesStatus.total === 0 && (
            <>
              <LoadingIndicator className="w-3.5 h-3.5 text-foreground-light shrink-0" />
              <span className="block truncate">Loading...</span>
            </>
          )}
        {panelDependenciesStatus.inputsAwaitingValue.length === 0 &&
          panelDependenciesStatus.total > 0 && (
            <>
              <LoadingIndicator className="w-3.5 h-3.5 text-foreground-light shrink-0" />
              <span className="block truncate">
                Loading{" "}
                {panelDependenciesStatus.status.complete.total +
                  panelDependenciesStatus.status.running.total +
                  1}{" "}
                of {panelDependenciesStatus.total + 1}...
              </span>
            </>
          )}
      </div>
    </BasePanelStatus>
  );
};

const PanelCancelled = ({ definition }) => {
  return (
    <BasePanelStatus definition={definition}>
      <div className="flex items-center space-x-1">
        <Icon
          className="w-3.5 h-3.5 text-foreground-light shrink-0"
          icon="cancel"
        />
        <span className="block truncate">Cancelled</span>
      </div>
    </BasePanelStatus>
  );
};

const PanelError = ({ definition }) => {
  return (
    <BasePanelStatus
      className="bg-alert-light border-alert text-foreground"
      definition={definition}
    >
      <div className="flex items-center space-x-1">
        <Icon
          className="w-3.5 h-3.5 text-alert shrink-0"
          icon="materialsymbols-solid:error"
        />
        <span className="block truncate">Error</span>
      </div>
      <span className="block">
        <ErrorMessage error={definition.error} />
      </span>
    </BasePanelStatus>
  );
};

const PanelStatus = ({ definition, showPanelError }: PanelStatusProps) => (
  <>
    {definition.status === "initialized" && (
      <PanelInitialized definition={definition} />
    )}
    {definition.status === "blocked" && (
      <PanelBlocked definition={definition} />
    )}
    {definition.status === "running" && (
      <PanelRunning definition={definition} />
    )}
    {definition.status === "cancelled" && (
      <PanelCancelled definition={definition} />
    )}
    {definition.status === "error" && showPanelError && (
      <PanelError definition={definition} />
    )}
  </>
);

export default PanelStatus;
