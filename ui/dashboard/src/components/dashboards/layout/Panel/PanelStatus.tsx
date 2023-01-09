import ErrorMessage from "../../../ErrorMessage";
import Icon from "../../../Icon";
import LoadingIndicator from "../../LoadingIndicator";
import { classNames } from "../../../../utils/styles";
import { HashLink } from "react-router-hash-link";
import { InputProperties } from "../../inputs/types";
import { PanelDefinition } from "../../../../types";
import { PanelDependencyStatuses } from "../../common/types";
import { ReactNode, useMemo } from "react";
import { useDashboard } from "../../../../hooks/useDashboard";
import { useLocation } from "react-router-dom";
import { usePanel } from "../../../../hooks/usePanel";

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

const usePanelDependenciesStatus = () => {
  const { dependenciesByStatus } = usePanel();
  const { selectedDashboardInputs } = useDashboard();
  return useMemo<PanelDependencyStatuses>(() => {
    const inputPanelsAwaitingValue: PanelDefinition[] = [];
    const initializedPanels: PanelDefinition[] = [];
    const blockedPanels: PanelDefinition[] = [];
    const runningPanels: PanelDefinition[] = [];
    const cancelledPanels: PanelDefinition[] = [];
    const errorPanels: PanelDefinition[] = [];
    const completePanels: PanelDefinition[] = [];
    let total = 0;
    for (const panels of Object.values(dependenciesByStatus)) {
      for (const panel of panels) {
        const isInput = panel.panel_type === "input";
        const inputProperties = isInput
          ? (panel.properties as InputProperties)
          : null;
        const hasInputValue =
          isInput &&
          inputProperties?.unqualified_name &&
          !!selectedDashboardInputs[inputProperties?.unqualified_name];
        total += 1;
        if (isInput && !hasInputValue) {
          inputPanelsAwaitingValue.push(panel);
        }
        if (panel.status === "initialized") {
          initializedPanels.push(panel);
        } else if (panel.status === "blocked") {
          blockedPanels.push(panel);
        } else if (panel.status === "running") {
          runningPanels.push(panel);
        } else if (panel.status === "cancelled") {
          completePanels.push(panel);
        } else if (panel.status === "error") {
          errorPanels.push(panel);
        } else if (panel.status === "complete") {
          completePanels.push(panel);
        }
      }
    }
    const status = {
      initialized: {
        total: initializedPanels.length,
        panels: initializedPanels,
      },
      blocked: {
        total: blockedPanels.length,
        panels: blockedPanels,
      },
      running: {
        total: runningPanels.length,
        panels: runningPanels,
      },
      cancelled: {
        total: cancelledPanels.length,
        panels: cancelledPanels,
      },
      error: {
        total: errorPanels.length,
        panels: errorPanels,
      },
      complete: {
        total: completePanels.length,
        panels: completePanels,
      },
    };
    return {
      total,
      inputsAwaitingValue: inputPanelsAwaitingValue,
      status,
    };
  }, [dependenciesByStatus]);
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
              Awaiting input value:{" "}
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
              <span className="block truncate">Running...</span>
            </>
          )}
        {panelDependenciesStatus.inputsAwaitingValue.length === 0 &&
          panelDependenciesStatus.total > 0 && (
            <>
              <LoadingIndicator className="w-3.5 h-3.5 text-foreground-light shrink-0" />
              <span className="block truncate">
                Running{" "}
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
              <span className="block truncate">Running...</span>
            </>
          )}
        {panelDependenciesStatus.inputsAwaitingValue.length === 0 &&
          panelDependenciesStatus.total > 0 && (
            <>
              <LoadingIndicator className="w-3.5 h-3.5 text-foreground-light shrink-0" />
              <span className="block truncate">
                Running{" "}
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
      className="bg-alert-light border-alert-light text-foreground"
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
