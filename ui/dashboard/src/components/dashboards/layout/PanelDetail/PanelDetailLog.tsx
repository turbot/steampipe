import DateTime from "../../../DateTime";
import Icon from "../../../Icon";
import Panel from "../Panel";
import sortBy from "lodash/sortBy";
import { classNames } from "../../../../utils/styles";
import { Disclosure } from "@headlessui/react";
import { PanelDetailProps } from "./index";
import { DashboardRunState, PanelLog } from "../../../../types";
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
    case "ready":
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
          icon="materialsymbols-solid:run-circle"
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
          icon="materialsymbols-solid:check-circle"
        />
      );
  }
};

const PanelLogStatus = ({ status }: PanelLogStatusProps) => {
  const baseClassname = "inline-block tabular-nums whitespace-nowrap";
  switch (status) {
    case "ready":
      return <pre className={baseClassname}>Ready&nbsp;&nbsp;&nbsp;</pre>;
    case "blocked":
      return <pre className={baseClassname}>Blocked&nbsp;</pre>;
    case "running":
      return <pre className={baseClassname}>Running&nbsp;</pre>;
    case "error":
      return <pre className={baseClassname}>Error&nbsp;&nbsp;&nbsp;</pre>;
    case "complete":
      return <pre className={baseClassname}>Complete</pre>;
  }
};

const PanelLogMessage = ({ log }: PanelLogRowProps) => (
  <div className="flex space-x-2">
    <PanelLogStatus status={log.status} />
    {log.prefix && (
      <span className="text-foreground-lighter">{log.prefix}:</span>
    )}
    {log.isDependency && <span className="">{log.title}</span>}
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
                        ? "materialsymbols-outline:expand-less"
                        : "materialsymbols-outline:expand-more"
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

const PanelLogs = () => {
  const { panelsLog } = useDashboard();
  const { definition, dependencies } = usePanel();
  const panelLog = panelsLog[definition.name];
  const dependencyPanelLogs: PanelLog[] = [];
  for (const dependency of dependencies || []) {
    const dependencyPanelLog = panelsLog[dependency.name];
    if (!dependencyPanelLog) {
      continue;
    }
    dependencyPanelLogs.push(
      ...dependencyPanelLog.map((l) => ({
        ...l,
        isDependency: true,
        prefix: "Dependency",
      }))
    );
  }
  const allLogs = sortBy([...dependencyPanelLogs, ...panelLog], "timestamp");
  console.log(dependencies);
  return (
    <div className="border border-black-scale-2 divide-y divide-divide">
      {allLogs.map((log) => (
        <PanelLogRow
          key={`${log.status}:${log.timestamp}:${log.prefix}:${log.title}`}
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
      title: `${definition.title ? ` ${definition.title} Log` : "Log"}`,
    }}
    showControls={false}
    showPanelError={false}
    forceBackground={true}
  >
    <PanelLogs />
  </Panel>
);

export default PanelDetailLog;
