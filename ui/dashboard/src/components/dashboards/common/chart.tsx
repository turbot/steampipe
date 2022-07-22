import { ChartTooltipFormatter } from "./types";
import { renderToString } from "react-dom/server";
import { classNames } from "../../../utils/styles";
import { ThemeProvider, ThemeWrapper, useTheme } from "../../../hooks/useTheme";

const Tooltip = ({ children, title }) => {
  return (
    <ThemeProvider>
      <ThemeWrapper>
        <div className="p-3 border border-divide rounded-md text-sm flex flex-col space-y-2">
          <Title title={title} />
          {children}
        </div>
      </ThemeWrapper>
    </ThemeProvider>
  );
};

const Title = ({ title }) => {
  const { theme } = useTheme();
  console.log(theme);
  return <strong className="block break-all">{title}</strong>;
};

const MetadataItem = ({ name, value }) => {
  return (
    <div>
      <span className="block text-sm text-table-head truncate">{name}</span>
      <span className={classNames("block", "break-words")}>{value}</span>
    </div>
    // <div className="space-x-2">
    //   <span>{name}</span>
    //   <span>=</span>
    //   <span>{value}</span>
    // </div>
  );
};

const Metadata = ({ metadata = {} }) => {
  return (
    <div className="space-y-2">
      {Object.entries(metadata).map(([key, value]) => (
        <MetadataItem key={key} name={key} value={value} />
      ))}
    </div>
  );
};

const formatChartTooltip = (params: any, data: any) => {
  const componentType = params.componentType;
  if (componentType !== "series") {
    return params.name;
  }
  const componentSubType = params.componentSubType;
  const dataType = params.dataType;

  switch (componentSubType) {
    case "graph":
      return new GraphTooltipFormatter().format(params, data);
  }
};

class GraphTooltipFormatter implements ChartTooltipFormatter {
  format(params, data: any[]): string {
    const dataRow = data[params.dataIndex];
    console.log({ params, data, dataRow });
    return renderToString(
      <Tooltip title={params.name}>
        <Metadata metadata={dataRow.metadata} />
      </Tooltip>
    );
  }
}

export { formatChartTooltip };
