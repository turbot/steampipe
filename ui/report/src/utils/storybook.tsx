import Report from "../components/reports/layout/Report";
import { ReportContext } from "../hooks/useReport";
import { noop } from "./func";

type PanelStoryDecoratorProps = {
  definition: any;
  nodeType: "card" | "chart" | "container" | "table" | "text";
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
        metadata: {
          mod: {
            full_name: "mod.storybook",
            short_name: "storybook",
          },
        },
        metadataLoaded: true,
        availableReportsLoaded: true,
        closePanelDetail: noop,
        dispatch: () => {},
        error: null,
        reports: [],
        selectedPanel: null,
        selectedReport: {
          title: "Storybook Report Wrapper",
          full_name: "storybook.report.storybook_report_wrapper",
          short_name: "storybook_report_wrapper",
          mod_full_name: "mod.storybook",
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
              sql: "storybook",
            },
          ],
        },
        sqlDataMap: {
          storybook: definition.data,
        },
      }}
    >
      <Report />
    </ReportContext.Provider>
  );
};
