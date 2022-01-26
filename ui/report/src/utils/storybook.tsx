import Report from "../components/reports/layout/Report";
import { ReportContext } from "../hooks/useReport";
import { noop } from "./func";

type PanelStoryDecoratorProps = {
  definition: any;
  nodeType: "container" | "chart" | "counter" | "table" | "text";
  additionalProperties?: {
    [key: string]: any;
  };
};

export const PanelStoryDecorator = ({
  definition = {},
  nodeType,
  additionalProperties = {},
}: PanelStoryDecoratorProps) => {
  const { properties, ...rest } = definition;

  return (
    <ReportContext.Provider
      value={{
        availableReportsLoaded: true,
        closePanelDetail: noop,
        dispatch: () => {},
        error: null,
        reports: [],
        selectedPanel: null,
        selectedReport: {
          name: "storybook.report.storybook_report_wrapper",
          title: "Storybook Report Wrapper",
        },
        report: {
          name: "storybook.report.storybook_report_wrapper",
          children: [
            {
              name: `${nodeType}.story`,
              node_type: nodeType,
              ...rest,
              properties: {
                ...(properties || {}),
                ...additionalProperties,
              },
            },
          ],
        },
      }}
    >
      <Report />
    </ReportContext.Provider>
  );
};
