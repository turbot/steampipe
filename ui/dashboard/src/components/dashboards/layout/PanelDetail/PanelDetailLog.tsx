import DateTime from "../../../DateTime";
import Icon from "../../../Icon";
import Panel from "../Panel";
import { classNames } from "../../../../utils/styles";
import { Disclosure } from "@headlessui/react";
import { PanelDetailProps } from "./index";
import { DashboardRunState, PanelLog } from "../../../../types";
import { useDashboard } from "../../../../hooks/useDashboard";

type PanelLogRowProps = {
  log: PanelLog;
};

type PanelLogsProps = {
  logs: PanelLog[];
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
          icon="materialsymbols-solid:pending"
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
  switch (status) {
    case "ready":
      return <span>Ready</span>;
    case "blocked":
      return <span>Blocked</span>;
    case "running":
      return <span>Running</span>;
    case "error":
      return <span>Error</span>;
    case "complete":
      return <span>Complete</span>;
  }
};

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
              <div className="flex items-center space-x-2">
                <PanelLogIcon status={log.status} />
                <DateTime date={log.timestamp} />
                <PanelLogStatus status={log.status} />
              </div>
              {log.error ? (
                <div>
                  <Icon
                    className="text-sm text-foreground-light"
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
              <Disclosure.Panel>{log.error}</Disclosure.Panel>
            ) : null}
          </>
        );
      }}
    </Disclosure>
  );
};

const PanelLogs = ({ logs }: PanelLogsProps) => {
  return (
    <div className="border border-black-scale-2 divide-y divide-divide">
      {logs.map((log) => (
        <PanelLogRow key={`${log.status}:${log.timestamp}`} log={log} />
      ))}
    </div>
  );
};

const PanelDetailLog = ({ definition }: PanelDetailProps) => {
  const { panelsLog } = useDashboard();
  const panelLog = panelsLog[definition.name];
  console.log(panelLog);
  return (
    <Panel definition={definition} showControls={false} forceBackground={true}>
      <PanelLogs logs={panelLog} />
    </Panel>
  );
};

export default PanelDetailLog;
