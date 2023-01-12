import DateTime from "../../../DateTime";
import Icon from "../../../Icon";
import Panel from "../Panel";
import sortBy from "lodash/sortBy";
import { classNames } from "../../../../utils/styles";
import {
  DashboardRunState,
  PanelDefinition,
  PanelLog,
  PanelsLog,
  PanelsMap,
} from "../../../../types";
import { Disclosure } from "@headlessui/react";
import { getNodeAndEdgeDataFormat } from "../../common/useNodeAndEdgeData";
import { NodeAndEdgeProperties } from "../../common/types";
import { PanelDetailProps } from "./index";
import { useDashboard } from "../../../../hooks/useDashboard";
import { usePanel } from "../../../../hooks/usePanel";

type PanelLogRowProps = {
  log: PanelLog;
};

type PanelLogIconProps = {
  status: DashboardRunState;
};

type PanelLogStatusProps = {
  status: DashboardRunState;
};

const PanelLogIcon = ({ status }: PanelLogIconProps) => {
  switch (status) {
    case "initialized":
      return (
        <Icon
          className="text-skip w-4.5 h-4.5"
          icon="materialsymbols-outline:pending"
        />
      );
    case "blocked":
      return (
        <Icon
          className="text-skip w-4.5 h-4.5"
          icon="materialsymbols-solid:block"
        />
      );
    case "running":
      return (
        <Icon
          className="text-skip w-4.5 h-4.5"
          icon="materialsymbols-solid:run_circle"
        />
      );
    case "cancelled":
      return (
        <Icon
          className="text-skip w-4.5 h-4.5"
          icon="materialsymbols-outline:cancel"
        />
      );
    case "error":
      return (
        <Icon
          className="text-alert w-4.5 h-4.5"
          icon="materialsymbols-solid:error"
        />
      );
    case "complete":
      return (
        <Icon
          className="text-ok w-4.5 h-4.5"
          icon="materialsymbols-solid:check_circle"
        />
      );
  }
};

const PanelLogStatus = ({ status }: PanelLogStatusProps) => {
  const baseClassname = "inline-block tabular-nums whitespace-nowrap";
  switch (status) {
    case "initialized":
      return <pre className={baseClassname}>Initialized</pre>;
    case "blocked":
      return (
        <pre className={baseClassname}>Blocked&nbsp;&nbsp;&nbsp;&nbsp;</pre>
      );
    case "running":
      return (
        <pre className={baseClassname}>Running&nbsp;&nbsp;&nbsp;&nbsp;</pre>
      );
    case "cancelled":
      return <pre className={baseClassname}>Cancelled&nbsp;&nbsp;</pre>;
    case "error":
      return (
        <pre className={baseClassname}>
          Error&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
        </pre>
      );
    case "complete":
      return <pre className={baseClassname}>Complete&nbsp;&nbsp;&nbsp;</pre>;
  }
};

const PanelLogMessage = ({ log }: PanelLogRowProps) => (
  <div className="flex space-x-2">
    <PanelLogStatus status={log.status} />
    {log.prefix && (
      <pre className="text-foreground-lighter tabular-nums">{log.prefix}:</pre>
    )}
    {log.isDependency && <span className="">{log.title}</span>}
    {log.executionTime !== undefined && (
      <span className="text-foreground-lighter tabular-nums">
        {log.executionTime.toLocaleString()}ms
      </span>
    )}
  </div>
);

const PanelLogRow = ({ log }: PanelLogRowProps) => {
  return (
    <Disclosure>
      {({ open }) => {
        return (
          <>
            <Disclosure.Button
              className={classNames(
                "w-full px-2 py-1 flex justify-between items-center hover:bg-black-scale-2",
                log.error ? "cursor-pointer" : "cursor-default"
              )}
            >
              <div className="flex items-center space-x-3">
                <PanelLogIcon status={log.status} />
                <DateTime
                  date={log.timestamp}
                  dateClassName="hidden"
                  timeFormat="HH:mm:ss.SSS"
                />
                <PanelLogMessage log={log} />
              </div>
              {log.error ? (
                <div>
                  <Icon
                    className="w-4.5 h-4.5 text-foreground-light"
                    icon={
                      open
                        ? "materialsymbols-outline:expand_less"
                        : "materialsymbols-outline:expand_more"
                    }
                  />
                </div>
              ) : null}
            </Disclosure.Button>
            {log.error ? (
              <Disclosure.Panel className="px-2 py-1">
                {log.error}
              </Disclosure.Panel>
            ) : null}
          </>
        );
      }}
    </Disclosure>
  );
};

const addDependencyLogs = (
  panel: PanelDefinition,
  panelsLog: PanelsLog,
  panelsMap: PanelsMap,
  dependencyLogs: PanelLog[],
  dependentPanelRecord
) => {
  for (const dependency of panel.dependencies || []) {
    if (dependentPanelRecord[dependency]) {
      continue;
    }
    dependentPanelRecord[dependency] = true;
    const dependencyPanel = panelsMap[dependency];
    if (!dependencyPanel) {
      continue;
    }
    addDependencyLogs(
      dependencyPanel,
      panelsLog,
      panelsMap,
      dependencyLogs,
      dependentPanelRecord
    );
  }
  const dependencyPanelLog = panelsLog[panel.name];
  dependencyLogs.push(
    ...dependencyPanelLog.map((l) => ({
      ...l,
      isDependency: true,
      prefix: panel.panel_type,
    }))
  );
};

const getDependencyLogs = (
  panel: PanelDefinition,
  panelsLog: PanelsLog,
  panelsMap: PanelsMap
) => {
  const dependencyLogs: PanelLog[] = [];
  const dependentPanelRecord = {};

  addDependencyLogs(
    panel,
    panelsLog,
    panelsMap,
    dependencyLogs,
    dependentPanelRecord
  );

  if (
    (panel.panel_type === "flow" ||
      panel.panel_type === "graph" ||
      panel.panel_type === "hierarchy") &&
    panel.properties &&
    getNodeAndEdgeDataFormat(panel.properties) === "NODE_AND_EDGE"
  ) {
    const nodeAndEdgeProperties = panel.properties as NodeAndEdgeProperties;
    for (const node of nodeAndEdgeProperties.nodes || []) {
      const nodePanel = panelsMap[node];
      if (!nodePanel) {
        continue;
      }
      addDependencyLogs(
        nodePanel,
        panelsLog,
        panelsMap,
        dependencyLogs,
        dependentPanelRecord
      );
    }
    for (const edge of nodeAndEdgeProperties.edges || []) {
      const edgePanel = panelsMap[edge];
      if (!edgePanel) {
        continue;
      }
      addDependencyLogs(
        edgePanel,
        panelsLog,
        panelsMap,
        dependencyLogs,
        dependentPanelRecord
      );
    }
  }
  return dependencyLogs;
};

const PanelLogs = () => {
  const { panelsLog, panelsMap } = useDashboard();
  const { definition } = usePanel();
  const panelLog = panelsLog[definition.name];
  const dependencyPanelLogs = getDependencyLogs(
    definition as PanelDefinition,
    panelsLog,
    panelsMap
  );
  const allLogs = sortBy([...dependencyPanelLogs, ...panelLog], "timestamp");
  return (
    <div className="border border-black-scale-2 divide-y divide-divide">
      {allLogs.map((log, idx) => (
        <PanelLogRow
          key={`${log.status}:${log.timestamp}:${log.prefix}:${log.title}-${idx}`}
          log={log}
        />
      ))}
    </div>
  );
};

const PanelDetailLog = ({ definition }: PanelDetailProps) => (
  <Panel
    definition={{
      ...definition,
      title: `${definition.title ? `${definition.title} Log` : "Log"}`,
    }}
    parentType="dashboard"
    showControls={false}
    showPanelError={false}
    forceBackground={true}
  >
    <PanelLogs />
  </Panel>
);

export default PanelDetailLog;
