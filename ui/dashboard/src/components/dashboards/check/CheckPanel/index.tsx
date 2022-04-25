import CheckSummaryChart from "../CheckSummaryChart";
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
  CheckNode,
  CheckResult,
  CheckResultStatus,
  CheckSummary,
} from "../common";
import { classNames } from "../../../../utils/styles";
import { ControlDimension } from "../Benchmark";
import { ThemeNames, useTheme } from "../../../../hooks/useTheme";
import { useMemo, useState } from "react";
import ControlResultNode from "../common/ControlResultNode";

interface CheckChildrenProps {
  children: CheckNode[];
  rootSummary: CheckSummary;
}

interface CheckResultsProps {
  results: CheckResult[];
  error?: string;
}

interface CheckPanelProps {
  node: CheckNode;
  rootSummary: CheckSummary;
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

const CheckChildren = ({ children, rootSummary }: CheckChildrenProps) => {
  if (!children) {
    return null;
  }

  return (
    <>
      {children.map((child) => (
        <CheckPanel key={child.name} node={child} rootSummary={rootSummary} />
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

const CheckResultRow = ({ result }: CheckResultRowProps) => {
  return (
    <div className="flex items-center bg-dashboard-panel p-4 last:rounded-b-md space-x-4">
      <div className="flex-shrink-0">
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
      <div className="flex-shrink-0 mr-4">
        <CheckResultRowStatusIcon status="error" />
      </div>
      <div className="flex-grow font-medium">{error}</div>
    </div>
  );
};

const CheckResults = ({ error, results }: CheckResultsProps) => {
  const { theme } = useTheme();

  if (!results) {
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
      {error && <CheckErrorRow error={error} />}
      {results.map((result) => (
        <CheckResultRow
          key={`${result.control.name}-${result.resource}`}
          result={result}
        />
      ))}
    </div>
  );
};

const CheckPanel = ({ node, rootSummary }: CheckPanelProps) => {
  const [expanded, setExpanded] = useState(false);
  const canBeExpanded =
    (!!node.children && node.children.length > 0) ||
    (!!node.results && node.results.length > 0) ||
    node.error;

  const [child_nodes, result_nodes] = useMemo(() => {
    const children: CheckNode[] = [];
    const results: CheckResult[] = [];
    for (const child of node.children || []) {
      if (child.type === "result") {
        results.push((child as ControlResultNode).result);
      } else {
        children.push(child);
      }
    }
    return [sortBy(children, "title"), results];
  }, [node]);

  return (
    <>
      <div id={node.name} className={getMargin(node.depth - 1)}>
        <section
          className={classNames(
            "bg-dashboard-panel cursor-pointer shadow-sm rounded-md",
            expanded && node.results && node.results.length > 0
              ? "rounded-b-none"
              : null,
            node.status !== "complete" && node.status !== "error"
              ? "animate-pulse"
              : null
          )}
          onClick={() => setExpanded((current) => !current)}
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
                  name={node.name}
                  summary={node.summary}
                  rootSummary={rootSummary}
                />
              </div>
            </div>
            {canBeExpanded && !expanded && (
              <ExpandCheckNodeIcon className="w-5 md:w-7 h-5 md:h-7 flex-shrink-0 text-foreground-lightest" />
            )}
            {expanded && (
              <CollapseBenchmarkIcon className="w-5 md:w-7 h-5 md:h-7 flex-shrink-0 text-foreground-lightest" />
            )}
            {!canBeExpanded && <div className="w-5 md:w-7 h-5 md:h-7" />}
          </div>
        </section>
        {expanded && <CheckResults results={result_nodes} />}
      </div>
      {expanded && (
        <CheckChildren children={child_nodes} rootSummary={rootSummary} />
      )}
    </>
  );
};

export default CheckPanel;
