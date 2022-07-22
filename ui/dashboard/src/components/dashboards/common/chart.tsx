import { ChartTooltipFormatter } from "./types";
import { classNames } from "../../../utils/styles";
import { renderToString } from "react-dom/server";
import { ThemeProvider, ThemeWrapper } from "../../../hooks/useTheme";

const Tooltip = ({ children, title }) => {
  return (
    <ThemeProvider>
      <ThemeWrapper>
        <div className="p-3 border border-divide rounded-md text-sm flex flex-col space-y-2 bg-dashboard-panel">
          <Title title={title} />
          {children}
        </div>
      </ThemeWrapper>
    </ThemeProvider>
  );
};

const Title = ({ title }) => {
  return <strong className="block break-all">{title}</strong>;
};

const PropertyItem = ({ name, value }) => {
  return (
    <div>
      <span className="block text-sm text-table-head truncate">{name}</span>
      {value === null && (
        <span className="text-foreground-lightest">
          <>null</>
        </span>
      )}
      {value !== null && (
        <span className={classNames("block", "break-words")}>{value}</span>
      )}
    </div>
    // <div className="space-x-2">
    //   <span>{name}</span>
    //   <span>=</span>
    //   <span>{value}</span>
    // </div>
  );
};

const Properties = ({ properties = {} }) => {
  return (
    <div className="space-y-2">
      {Object.entries(properties || {}).map(([key, value]) => (
        <PropertyItem key={key} name={key} value={value} />
      ))}
    </div>
  );
};

const formatChartTooltip = (params: any) => {
  const componentType = params.componentType;
  if (componentType !== "series") {
    return params.name;
  }
  const componentSubType = params.componentSubType;
  const dataType = params.dataType;

  switch (componentSubType) {
    case "graph":
      return new GraphTooltipFormatter().format(params);
  }
};

class GraphTooltipFormatter implements ChartTooltipFormatter {
  format(params): string {
    const data = params.data;
    const tooltip = renderToString(
      <Tooltip title={params.name}>
        <Properties properties={data.properties} />
      </Tooltip>
    );
    return tooltip;
  }
}

export { formatChartTooltip };
