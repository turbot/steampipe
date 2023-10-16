import CheckSummaryChart from "../CheckSummaryChart";
import ControlDimension from "../Benchmark/ControlDimension";
import ControlEmptyResultNode from "../common/node/ControlEmptyResultNode";
import ControlErrorNode from "../common/node/ControlErrorNode";
import ControlResultNode from "../common/node/ControlResultNode";
import sortBy from "lodash/sortBy";
import {
  AlarmIcon,
  CollapseBenchmarkIcon,
  EmptyIcon,
  ErrorIcon,
  ExpandCheckNodeIcon,
  InfoIcon,
  OKIcon,
  SkipIcon,
  UnknownIcon,
} from "../../../../constants/icons";
import {
  CheckGroupingActions,
  useCheckGrouping,
} from "../../../../hooks/useCheckGrouping";
import {
  CheckNode,
  CheckResult,
  CheckResultStatus,
  CheckSeveritySummary,
} from "../common";
import { classNames } from "../../../../utils/styles";
import { useMemo } from "react";

type CheckChildrenProps = {
  depth: number;
  children: CheckNode[];
};

type CheckResultsProps = {
  empties: ControlEmptyResultNode[];
  errors: ControlErrorNode[];
  results: ControlResultNode[];
};

type CheckPanelProps = {
  depth: number;
  node: CheckNode;
};

type CheckPanelSeverityProps = {
  severity_summary: CheckSeveritySummary;
};

type CheckPanelSeverityBadgeProps = {
  label: string;
  count: number;
  title: string;
};

type CheckEmptyResultRowProps = {
  node: ControlEmptyResultNode;
};

type CheckResultRowProps = {
  result: CheckResult;
};

type CheckErrorRowProps = {
  error: string;
};

type CheckResultRowStatusIconProps = {
  status: CheckResultStatus;
};

const getMargin = (depth) => {
  switch (depth) {
    case 1:
      return "ml-[6px] md:ml-[24px]";
    case 2:
      return "ml-[12px] md:ml-[48px]";
    case 3:
      return "ml-[18px] md:ml-[72px]";
    case 4:
      return "ml-[24px] md:ml-[96px]";
    case 5:
      return "ml-[30px] md:ml-[120px]";
    case 6:
      return "ml-[36px] md:ml-[144px]";
    default:
      return "ml-0";
  }
};

const CheckChildren = ({ children, depth }: CheckChildrenProps) => {
  if (!children) {
    return null;
  }

  return (
    <>
      {children.map((child) => (
        <CheckPanel key={child.name} depth={depth} node={child} />
      ))}
    </>
  );
};

const CheckResultRowStatusIcon = ({
  status,
}: CheckResultRowStatusIconProps) => {
  switch (status) {
    case CheckResultStatus.alarm:
      return <AlarmIcon className="h-5 w-5 text-alert" />;
    case CheckResultStatus.error:
      return <ErrorIcon className="h-5 w-5 text-alert" />;
    case CheckResultStatus.ok:
      return <OKIcon className="h-5 w-5 text-ok" />;
    case CheckResultStatus.info:
      return <InfoIcon className="h-5 w-5 text-info" />;
    case CheckResultStatus.skip:
      return <SkipIcon className="h-5 w-5 text-skip" />;
    case CheckResultStatus.empty:
      return <EmptyIcon className="h-5 w-5 text-skip" />;
    default:
      return <UnknownIcon className="h-5 w-5 text-skip" />;
  }
};

const getCheckResultRowIconTitle = (status: CheckResultStatus) => {
  switch (status) {
    case CheckResultStatus.error:
      return "Error";
    case CheckResultStatus.alarm:
      return "Alarm";
    case CheckResultStatus.ok:
      return "OK";
    case CheckResultStatus.info:
      return "Info";
    case CheckResultStatus.skip:
      return "Skipped";
    case CheckResultStatus.empty:
      return "No results";
  }
};

const CheckResultRow = ({ result }: CheckResultRowProps) => {
  return (
    <div className="flex bg-dashboard-panel print:bg-white p-4 last:rounded-b-md space-x-4">
      <div
        className="flex-shrink-0"
        title={getCheckResultRowIconTitle(result.status)}
      >
        <CheckResultRowStatusIcon status={result.status} />
      </div>
      <div className="flex flex-col md:flex-row flex-grow">
        <div className="md:flex-grow leading-4 mt-px">{result.reason}</div>
        <div className="flex space-x-2 mt-2 md:mt-px md:text-right">
          {(result.dimensions || []).map((dimension) => (
            <ControlDimension
              key={dimension.key}
              dimensionKey={dimension.key}
              dimensionValue={dimension.value}
            />
          ))}
        </div>
      </div>
    </div>
  );
};

const CheckEmptyResultRow = ({ node }: CheckEmptyResultRowProps) => {
  return (
    <div className="flex bg-dashboard-panel print:bg-white p-4 last:rounded-b-md space-x-4">
      <div
        className="flex-shrink-0"
        title={getCheckResultRowIconTitle(CheckResultStatus.empty)}
      >
        <CheckResultRowStatusIcon status={CheckResultStatus.empty} />
      </div>
      <div className="leading-4 mt-px">{node.title}</div>
    </div>
  );
};

const CheckErrorRow = ({ error }: CheckErrorRowProps) => {
  return (
    <div className="flex bg-dashboard-panel print:bg-white p-4 last:rounded-b-md space-x-4">
      <div
        className="flex-shrink-0"
        title={getCheckResultRowIconTitle(CheckResultStatus.error)}
      >
        <CheckResultRowStatusIcon status={CheckResultStatus.error} />
      </div>
      <div className="leading-4 mt-px">{error}</div>
    </div>
  );
};

const CheckResults = ({ empties, errors, results }: CheckResultsProps) => {
  if (empties.length === 0 && errors.length === 0 && results.length === 0) {
    return null;
  }

  return (
    <div
      className={classNames(
        "border-t shadow-sm rounded-b-md divide-y divide-table-divide border-divide print:shadow-none print:border print:break-before-avoid-page print:break-after-avoid-page print:break-inside-auto",
      )}
    >
      {empties.map((emptyNode) => (
        <CheckEmptyResultRow key={`${emptyNode.name}`} node={emptyNode} />
      ))}
      {errors.map((errorNode) => (
        <CheckErrorRow key={`${errorNode.name}`} error={errorNode.error} />
      ))}
      {results.map((resultNode) => (
        <CheckResultRow
          key={`${resultNode.result.control.name}-${
            resultNode.result.resource
          }${
            resultNode.result.dimensions
              ? `-${resultNode.result.dimensions
                  .map((d) => `${d.key}=${d.value}`)
                  .join("-")}`
              : ""
          }`}
          result={resultNode.result}
        />
      ))}
    </div>
  );
};

const CheckPanelSeverityBadge = ({
  count,
  label,
  title,
}: CheckPanelSeverityBadgeProps) => {
  return (
    <div
      className={classNames(
        "border rounded-md text-sm divide-x",
        count > 0 ? "border-severity" : "border-skip",
        count > 0
          ? "bg-severity text-white divide-white"
          : "text-skip divide-skip",
      )}
      title={title}
    >
      <span className={classNames("px-2 py-px")}>{label}</span>
      {count > 0 && <span className={classNames("px-2 py-px")}>{count}</span>}
    </div>
  );
};

const CheckPanelSeverity = ({ severity_summary }: CheckPanelSeverityProps) => {
  const critical = severity_summary["critical"];
  const high = severity_summary["high"];

  if (critical === undefined && high === undefined) {
    return null;
  }

  return (
    <>
      {critical !== undefined && (
        <CheckPanelSeverityBadge
          label="Critical"
          count={critical}
          title={`${critical.toLocaleString()} critical severity ${
            critical === 1 ? "result" : "results"
          }`}
        />
      )}
      {high !== undefined && (
        <CheckPanelSeverityBadge
          label="High"
          count={high}
          title={`${high.toLocaleString()} high severity ${
            high === 1 ? "result" : "results"
          }`}
        />
      )}
    </>
  );
};

const CheckPanel = ({ depth, node }: CheckPanelProps) => {
  const { firstChildSummaries, dispatch, groupingsConfig, nodeStates } =
    useCheckGrouping();
  const expanded = nodeStates[node.name]
    ? nodeStates[node.name].expanded
    : false;

  const [child_nodes, error_nodes, empty_nodes, result_nodes, can_be_expanded] =
    useMemo(() => {
      const children: CheckNode[] = [];
      const errors: ControlErrorNode[] = [];
      const empty: ControlEmptyResultNode[] = [];
      const results: ControlResultNode[] = [];
      for (const child of node.children || []) {
        if (child.type === "error") {
          errors.push(child as ControlErrorNode);
        } else if (child.type === "result") {
          results.push(child as ControlResultNode);
        } else if (child.type === "empty_result") {
          empty.push(child as ControlEmptyResultNode);
        } else if (child.type !== "running") {
          children.push(child);
        }
      }
      return [
        sortBy(children, "sort"),
        sortBy(errors, "sort"),
        sortBy(empty, "sort"),
        results,
        children.length > 0 ||
          (groupingsConfig &&
            groupingsConfig.length > 0 &&
            groupingsConfig[groupingsConfig.length - 1].type === "result" &&
            (errors.length > 0 || empty.length > 0 || results.length > 0)),
      ];
    }, [groupingsConfig, node]);

  return (
    <>
      <div
        id={node.name}
        className={classNames(
          getMargin(depth - 1),
          depth === 1 && node.type === "benchmark"
            ? "print:break-before-page"
            : null,
          node.type === "benchmark" || node.type === "control"
            ? "print:break-inside-avoid-page"
            : null,
        )}
      >
        <section
          className={classNames(
            "bg-dashboard-panel shadow-sm rounded-md border-divide print:border print:bg-white print:shadow-none",
            can_be_expanded ? "cursor-pointer" : null,
            expanded &&
              (empty_nodes.length > 0 ||
                error_nodes.length > 0 ||
                result_nodes.length > 0)
              ? "rounded-b-none border-b-0"
              : null,
          )}
          onClick={() =>
            can_be_expanded
              ? dispatch({
                  type: expanded
                    ? CheckGroupingActions.COLLAPSE_NODE
                    : CheckGroupingActions.EXPAND_NODE,
                  name: node.name,
                })
              : null
          }
        >
          <div className="p-4 flex items-center space-x-6">
            <div className="flex flex-grow justify-between items-center space-x-6">
              <div className="flex items-center space-x-4">
                <h3
                  id={`${node.name}-title`}
                  className="mt-0"
                  title={node.title}
                >
                  {node.title}
                </h3>
                <CheckPanelSeverity severity_summary={node.severity_summary} />
              </div>
              <div className="flex-shrink-0 w-40 md:w-72 lg:w-96">
                <CheckSummaryChart
                  status={node.status}
                  summary={node.summary}
                  firstChildSummaries={firstChildSummaries}
                />
              </div>
            </div>
            {can_be_expanded && !expanded && (
              <ExpandCheckNodeIcon className="w-5 md:w-7 h-5 md:h-7 flex-shrink-0 text-foreground-lightest" />
            )}
            {expanded && (
              <CollapseBenchmarkIcon className="w-5 md:w-7 h-5 md:h-7 flex-shrink-0 text-foreground-lightest" />
            )}
            {!can_be_expanded && <div className="w-5 md:w-7 h-5 md:h-7" />}
          </div>
        </section>
        {can_be_expanded &&
          expanded &&
          groupingsConfig &&
          groupingsConfig[groupingsConfig.length - 1].type === "result" && (
            <CheckResults
              empties={empty_nodes}
              errors={error_nodes}
              results={result_nodes}
            />
          )}
      </div>
      {can_be_expanded && expanded && (
        <CheckChildren children={child_nodes} depth={depth + 1} />
      )}
    </>
  );
};

export default CheckPanel;
