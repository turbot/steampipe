import CheckSummaryChart from "../CheckSummaryChart";
import { CheckNode, CheckResult, CheckResultStatus } from "../common";
import { classNames } from "../../../../utils/styles";
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
import { ThemeNames, useTheme } from "../../../../hooks/useTheme";
import { useState } from "react";
import { ControlDimension } from "../Benchmark";

interface CheckPanelProps {
  node: CheckNode;
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
      return "ml-[24px]";
    case 2:
      return "ml-[48px]";
    case 3:
      return "ml-[72px]";
    case 4:
      return "ml-[96px]";
    case 5:
      return "ml-[120px]";
    case 6:
      return "ml-[144px]";
    default:
      return "ml-0";
  }
};

const CheckChildren = ({ node }: CheckPanelProps) => {
  if (!node.children) {
    return null;
  }

  return (
    <>
      {node.children.map((child) => (
        <CheckPanel key={child.name} node={child} />
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
    <div className="flex items-center bg-dashboard-panel p-4 last:rounded-b-md">
      <div className="flex-shrink-0 mr-4">
        <CheckResultRowStatusIcon status={result.status} />
      </div>
      <div className="flex-grow font-medium">{result.reason}</div>
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

const CheckResults = ({ node }: CheckPanelProps) => {
  const { theme } = useTheme();

  if (!node.results) {
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
      {node.error && <CheckErrorRow error={node.error} />}
      {node.results.map((result) => (
        <CheckResultRow key={result.resource} result={result} />
      ))}
    </div>
  );
};

const CheckPanel = ({ node }: CheckPanelProps) => {
  const [expanded, setExpanded] = useState(false);

  return (
    <>
      <div id={node.name} className={getMargin(node.depth - 1)}>
        <section
          className={classNames(
            "bg-dashboard-panel cursor-pointer shadow-sm rounded-md",
            expanded && node.results ? "rounded-b-none" : null
          )}
          onClick={() => setExpanded((current) => !current)}
        >
          <div className="p-4 flex items-center space-x-6">
            <div className="flex flex-grow justify-between items-center">
              <h3
                id={`${node.name}-title`}
                className="truncate mt-0"
                title={node.title}
              >
                {node.title}
              </h3>
              <div className="w-48 md:w-72 lg:w-96">
                <CheckSummaryChart name={node.name} summary={node.summary} />
              </div>
            </div>
            {!expanded && (
              <ExpandCheckNodeIcon className="h-7 w-7 flex-shrink-0 text-foreground-lightest" />
            )}
            {expanded && (
              <CollapseBenchmarkIcon className="h-7 w-7 flex-shrink-0 text-foreground-lightest" />
            )}
          </div>
        </section>
        {expanded && <CheckResults node={node} />}
      </div>
      {expanded && <CheckChildren node={node} />}
    </>
  );
};

export default CheckPanel;
