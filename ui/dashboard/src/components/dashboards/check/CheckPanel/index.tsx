import CheckSummaryChart from "../CheckSummaryChart";
import ControlErrorNode from "../common/node/ControlErrorNode";
import ControlResultNode from "../common/node/ControlResultNode";
import sortBy from "lodash/sortBy";
import {
  AlarmIcon,
  CollapseBenchmarkIcon,
  ErrorIcon,
  ExpandCheckNodeIcon,
  InfoIcon,
  OKIcon,
  SkipIcon,
  UnknownIcon,
} from "../../../../constants/icons";
import {
  CheckDisplayGroup,
  CheckNode,
  CheckResult,
  CheckResultStatus,
  CheckSeverity,
  CheckSeveritySummary,
  CheckSummary,
} from "../common";
import { classNames } from "../../../../utils/styles";
import { ControlDimension } from "../Benchmark";
import { ThemeNames, useTheme } from "../../../../hooks/useTheme";
import { useMemo, useState } from "react";

interface CheckChildrenProps {
  depth: number;
  children: CheckNode[];
  groupingConfig: CheckDisplayGroup[];
  firstChildSummaries: CheckSummary[];
}

interface CheckResultsProps {
  results: ControlResultNode[];
  errors: ControlErrorNode[];
}

interface CheckPanelProps {
  depth: number;
  node: CheckNode;
  groupingConfig: CheckDisplayGroup[];
  firstChildSummaries: CheckSummary[];
}

interface CheckPanelSeverityProps {
  severity_summary: CheckSeveritySummary;
}

interface CheckPanelSeverityBadgeProps {
  label: string;
  count: number;
}

interface CheckResultRowProps {
  result: CheckResult;
}

interface CheckErrorRowProps {
  error: string;
}

interface CheckResultRowStatusIconProps {
  status: CheckResultStatus;
}

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

const CheckChildren = ({
  children,
  depth,
  groupingConfig,
  firstChildSummaries,
}: CheckChildrenProps) => {
  if (!children) {
    return null;
  }

  return (
    <>
      {children.map((child) => (
        <CheckPanel
          key={child.name}
          depth={depth}
          node={child}
          groupingConfig={groupingConfig}
          firstChildSummaries={firstChildSummaries}
        />
      ))}
    </>
  );
};

const CheckResultRowStatusIcon = ({
  status,
}: CheckResultRowStatusIconProps) => {
  switch (status) {
    case "alarm":
      return <AlarmIcon className="h-5 w-5 text-alert" />;
    case "error":
      return <ErrorIcon className="h-5 w-5 text-alert" />;
    case "ok":
      return <OKIcon className="h-5 w-5 text-ok" />;
    case "info":
      return <InfoIcon className="h-5 w-5 text-info" />;
    case "skip":
      return <SkipIcon className="h-5 w-5 text-skip" />;
    default:
      return <UnknownIcon className="h-5 w-5 text-skip" />;
  }
};

const getCheckResultRowIconTitle = (status: CheckResultStatus) => {
  switch (status) {
    case "error":
      return "Error";
    case "alarm":
      return "Alarm";
    case "ok":
      return "OK";
    case "info":
      return "Info";
    case "skip":
      return "Skipped";
  }
};

const CheckResultRow = ({ result }: CheckResultRowProps) => {
  return (
    <div className="flex bg-dashboard-panel p-4 last:rounded-b-md space-x-4">
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

const CheckErrorRow = ({ error }: CheckErrorRowProps) => {
  return (
    <div className="flex bg-dashboard-panel p-4 last:rounded-b-md space-x-4">
      <div
        className="flex-shrink-0"
        title={getCheckResultRowIconTitle("error")}
      >
        <CheckResultRowStatusIcon status="error" />
      </div>
      <div className="leading-4 mt-px">{error}</div>
    </div>
  );
};

const CheckResults = ({ errors, results }: CheckResultsProps) => {
  const { theme } = useTheme();

  if (!errors || !results) {
    return null;
  }

  return (
    <div
      className={classNames(
        "border-t shadow-sm rounded-b-md divide-y divide-table-divide",
        theme.name === ThemeNames.STEAMPIPE_DARK
          ? "border-table-divide"
          : "border-background"
      )}
    >
      {errors.map((errorNode) => (
        <CheckErrorRow key={`${errorNode.name}`} error={errorNode.error} />
      ))}
      {results.map((resultNode) => (
        <CheckResultRow
          key={`${resultNode.result.control.name}-${resultNode.result.resource}`}
          result={resultNode.result}
        />
      ))}
    </div>
  );
};

const CheckPanelSeverityBadge = ({
  label,
  count,
}: CheckPanelSeverityBadgeProps) => {
  return (
    <div
      className={classNames(
        "border rounded-md text-sm divide-x",
        count > 0 ? "border-yellow" : "border-skip",
        count > 0
          ? "bg-yellow text-white divide-white"
          : "text-skip divide-skip"
      )}
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
        <CheckPanelSeverityBadge label="Critical" count={critical} />
      )}
      {high !== undefined && (
        <CheckPanelSeverityBadge label="High" count={high} />
      )}
    </>
  );
};

const CheckPanel = ({
  depth,
  node,
  groupingConfig,
  firstChildSummaries,
}: CheckPanelProps) => {
  const [expanded, setExpanded] = useState(false);

  const [child_nodes, error_nodes, result_nodes, can_be_expanded] =
    useMemo(() => {
      const children: CheckNode[] = [];
      const errors: ControlErrorNode[] = [];
      const results: ControlResultNode[] = [];
      for (const child of node.children || []) {
        if (child.type === "error") {
          errors.push(child as ControlErrorNode);
        } else if (child.type === "result") {
          results.push(child as ControlResultNode);
        } else if (child.type !== "running") {
          children.push(child);
        }
      }
      return [
        sortBy(children, "sort"),
        errors,
        results,
        children.length > 0 ||
          (groupingConfig &&
            groupingConfig[groupingConfig.length - 1].type === "result" &&
            (errors.length > 0 || results.length > 0)),
      ];
    }, [groupingConfig, node]);

  return (
    <>
      <div id={node.name} className={getMargin(depth - 1)}>
        <section
          className={classNames(
            "bg-dashboard-panel shadow-sm rounded-md",
            can_be_expanded ? "cursor-pointer" : null,
            expanded && (error_nodes.length > 0 || result_nodes.length > 0)
              ? "rounded-b-none"
              : null
          )}
          onClick={() =>
            can_be_expanded ? setExpanded((current) => !current) : null
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
        {expanded &&
          groupingConfig &&
          groupingConfig[groupingConfig.length - 1].type === "result" && (
            <CheckResults errors={error_nodes} results={result_nodes} />
          )}
      </div>
      {expanded && (
        <CheckChildren
          children={child_nodes}
          depth={depth + 1}
          groupingConfig={groupingConfig}
          firstChildSummaries={firstChildSummaries}
        />
      )}
    </>
  );
};

export default CheckPanel;
