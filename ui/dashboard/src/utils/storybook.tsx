import Dashboard from "../components/dashboards/layout/Dashboard";
import { DashboardContext } from "../hooks/useDashboard";
import { noop } from "./func";

type PanelStoryDecoratorProps = {
  definition: any;
  panelType: "card" | "chart" | "container" | "table" | "text";
  additionalProperties?: {
    [key: string]: any;
  };
};

export const PanelStoryDecorator = ({
  definition = {},
  panelType,
  additionalProperties = {},
}: PanelStoryDecoratorProps) => {
  const { properties, ...rest } = definition;

  // const newPanel = {
  //   name: `${panelType}.story`,
  //   panel_type: panelType,
  //   ...rest,
  //   properties: {
  //     ...(properties || {}),
  //     ...additionalProperties,
  //   },
  //   sql: "storybook",
  // };

  return (
    <DashboardContext.Provider
      value={{
        dataMode: "live",
        snapshotId: null,
        dispatch: () => {},
        lastChangedInput: null,

        selectedSnapshot: null,
        refetchDashboard: false,
      }}
    >
      <Dashboard />
    </DashboardContext.Provider>
  );
};
