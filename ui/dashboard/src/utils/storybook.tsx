import Dashboard from "../components/dashboards/layout/Dashboard";
import { DashboardContext, DashboardSearch } from "../hooks/useDashboard";
import { noop } from "./func";

type PanelStoryDecoratorProps = {
  definition: any;
  nodeType: "card" | "chart" | "container" | "table" | "text";
  additionalProperties?: {
    [key: string]: any;
  };
};

const stubDashboardSearch: DashboardSearch = {
  value: "",
  groupBy: { value: "mod", tag: null },
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
          telemetry: "none",
        },
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
          tags: {},
          mod_full_name: "mod.storybook",
        },
        selectedDashboardInputs: {},
        lastChangedInput: null,
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
          dashboard: "storybook.dashboard.storybook_dashboard_wrapper",
        },

        sqlDataMap: {
          storybook: definition.data,
        },

        dashboardTags: {
          keys: [],
        },

        search: stubDashboardSearch,

        breakpointContext: {
          currentBreakpoint: "xl",
          maxBreakpoint: () => true,
          minBreakpoint: () => true,
          width: 0,
        },

        themeContext: {
          theme: {
            label: "Steampipe Default",
            name: "steampipe-default",
          },
          setTheme: noop,
          wrapperRef: null,
        },
      }}
    >
      <Dashboard />
    </DashboardContext.Provider>
  );
};
