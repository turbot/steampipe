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
      return "md:ml-[24px]";
    case 2:
      return "md:ml-[48px]";
    case 3:
      return "md:ml-[72px]";
    case 4:
      return "md:ml-[96px]";
    case 5:
      return "md:ml-[120px]";
    case 6:
      return "md:ml-[144px]";
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
      return <SkipIcon className="h-5 w-5 text-tbd" />;
    default:
      return <UnknownIcon className="h-5 w-5 text-tbd" />;
  }
};

const getCheckResultRowIconTitle = (status: CheckResultStatus) => {
  switch (status) {
    case "error":
      return "Control in error status";
    case "alarm":
      return "Control in alarm status";
    case "ok":
      return "Control in OK status";
    case "info":
      return "Control in info status";
    case "skip":
      return "Control in skip status";
  }
};

const CheckResultRow = ({ result }: CheckResultRowProps) => {
  return (
    <div className="flex items-center bg-dashboard-panel p-4 last:rounded-b-md space-x-4">
      <div
        className="flex-shrink-0"
        title={getCheckResultRowIconTitle(result.status)}
      >
        <CheckResultRowStatusIcon status={result.status} />
      </div>
      <div className="flex-grow">{result.reason}</div>
      <div className="flex-wrap space-x-2">
        {(result.dimensions || []).map((dimension) => (
          <ControlDimension
            key={dimension.key}
            dimensionKey={dimension.key}
            dimensionValue={dimension.value}
          />
        ))}
      </div>
    </div>
  );
};

const CheckErrorRow = ({ error }: CheckErrorRowProps) => {
  return (
    <div className="flex items-center bg-dashboard-panel p-4 last:rounded-b-md">
      <div
        className="flex-shrink-0 mr-4"
        title={getCheckResultRowIconTitle("error")}
      >
        <CheckResultRowStatusIcon status="error" />
      </div>
      <div className="flex-grow">{error}</div>
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
        } else {
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
              : null,
            node.status !== "complete" && node.status !== "error"
              ? "animate-pulse"
              : null
          )}
          onClick={() =>
            can_be_expanded ? setExpanded((current) => !current) : null
          }
        >
          <div className="p-4 flex items-center space-x-6">
            <div className="flex flex-grow justify-between items-center space-x-6">
              <div>
                <h3
                  id={`${node.name}-title`}
                  className="mt-0"
                  title={node.title}
                >
                  {node.title}
                </h3>
              </div>
              <div className="flex-shrink-0 w-40 md:w-72 lg:w-96">
                <CheckSummaryChart
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
