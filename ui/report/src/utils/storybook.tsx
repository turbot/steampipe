import Dashboard from "../components/dashboards/layout/Dashboard";
import { DashboardContext } from "../hooks/useDashboard";
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
    <DashboardContext.Provider
      value={{
        metadata: {
          mod: {
            title: "Storybook",
            full_name: "mod.storybook",
            short_name: "storybook",
          },
          installed_mods: {},
        },
        metadataLoaded: true,
        availableDashboardsLoaded: true,
        closePanelDetail: noop,
        dispatch: () => {},
        error: null,
        dashboards: [],
        selectedPanel: null,
        selectedDashboard: {
          title: "Storybook Dashboard Wrapper",
          full_name: "storybook.dashboard.storybook_dashboard_wrapper",
          short_name: "storybook_dashboard_wrapper",
          mod_full_name: "mod.storybook",
        },
        dashboard: {
          name: "storybook.dashboard.storybook_dashboard_wrapper",
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
      <Dashboard />
    </DashboardContext.Provider>
  );
};
