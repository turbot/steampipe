import ErrorMessage from "../../../ErrorMessage";
import Icon from "../../../Icon";
import { classNames } from "../../../../utils/styles";
import { ErrorRow, PendingRow } from "../../common/NodeAndEdgePanelInformation";
import { DashboardInputs, PanelDefinition } from "../../../../types";
import { InputProperties } from "../../inputs/types";
import { PanelDependencyStatuses } from "../../common/types";
import { ReactNode, useMemo } from "react";
import { usePanel } from "../../../../hooks/usePanel";
import { useDashboard } from "../../../../hooks/useDashboard";

type PanelStatusProps = PanelStatusBaseProps & {
  definition: PanelDefinition;
  showPanelError: boolean;
};

type PanelStatusBaseProps = {
  definition: PanelDefinition;
};

type PanelErrorProps = PanelStatusBaseProps & {
  error: string;
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
      "flex w-full h-full p-4 break-keep border items-center shadow rounded-md",
      !!definition.title ? "rounded-t-none" : null
    )}
  >
    {children}
  </div>
);

type CompleteRowProps = {
  definition: PanelDefinition;
  inputs: DashboardInputs;
  title: string;
};

const CompleteRow = ({ definition, inputs, title }: CompleteRowProps) => {
  const isInput = definition.panel_type === "input";
  const inputProperties = isInput
    ? (definition.properties as InputProperties)
    : null;
  const displayType = definition.display_type;
  const hasInputValue =
    isInput &&
    inputProperties?.unqualified_name &&
    !!inputs[inputProperties?.unqualified_name];

  return (
    <div className="flex items-center space-x-1">
      <Icon
        className="w-3.5 h-3.5 text-ok shrink-0"
        icon="materialsymbols-solid:check_circle"
      />
      <span className="block space-x-2 truncate">
        {title}
        {isInput && !hasInputValue ? (
          <span className="italic text-foreground-light">
            {" "}
            {!displayType && "Please select a value"}
            {displayType === "select" && "Please select a value"}
            {displayType === "multiselect" && "Please select a value"}
            {displayType === "combo" && "Please select a value"}
            {displayType === "multicombo" && "Please select a value"}
            {displayType === "table" && "Please select a value"}
            {displayType === "text" && "Please enter a value"}
          </span>
        ) : null}
      </span>
    </div>
  );
};

const PanelInitialized = ({ definition }: PanelStatusBaseProps) => {
  return <BasePanelStatus definition={definition}>Initialized</BasePanelStatus>;
};

const PanelBlocked = ({ definition }) => {
  const { selectedDashboardInputs } = useDashboard();
  const { dependenciesByStatus } = usePanel();
  const statuses = useMemo<PanelDependencyStatuses>(() => {
    const initializedPanels: PanelDefinition[] = [];
    const blockedPanels: PanelDefinition[] = [];
    const runningPanels: PanelDefinition[] = [];
    const cancelledPanels: PanelDefinition[] = [];
    const errorPanels: PanelDefinition[] = [];
    const completePanels: PanelDefinition[] = [];
    for (const panels of Object.values(dependenciesByStatus)) {
      for (const panel of panels) {
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
    return {
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
  }, [dependenciesByStatus]);
  return (
    <BasePanelStatus definition={definition}>
      <div className="space-y-1">
        {/*<div>{JSON.stringify(statuses)}</div>*/}
        {statuses.complete.total} complete, {statuses.running.total} running,{" "}
        {statuses.error.total} {statuses.error.total === 1 ? "error" : "errors"}
        {statuses.running.panels.map((panel) => (
          <PendingRow
            key={panel.name}
            title={`${panel.panel_type}: ${panel.title || panel.name}`}
          />
        ))}
        {statuses.error.panels.map((panel) => (
          <ErrorRow
            key={panel.name}
            title={`${panel.panel_type}: ${panel.title || panel.name}`}
            error={panel.error}
          />
        ))}
        {statuses.complete.panels.map((panel) => (
          <CompleteRow
            key={panel.name}
            definition={panel}
            inputs={selectedDashboardInputs}
            title={`${panel.panel_type}: ${panel.title || panel.name}`}
          />
        ))}
      </div>
    </BasePanelStatus>
  );
};

const PanelRunning = ({ definition }) => {
  return <BasePanelStatus definition={definition}>Running</BasePanelStatus>;
};

const PanelCancelled = ({ definition }) => {
  return <BasePanelStatus definition={definition}>Cancelled</BasePanelStatus>;
};

const PanelError = ({ definition, error }: PanelErrorProps) => {
  return (
    <BasePanelStatus
      className="bg-alert-light border-alert-light text-foreground"
      definition={definition}
    >
      <ErrorMessage error={error} />
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
    {definition.status === "error" && !!definition.error && showPanelError && (
      <PanelError definition={definition} error={definition.error as string} />
    )}
  </>
);

export default PanelStatus;
