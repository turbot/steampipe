import "../utils/registerComponents";
import Dashboard from "../components/dashboards/layout/Dashboard";
import { buildComponentsMap } from "../components";
import { DashboardContext } from "../hooks/useDashboard";
import {
  DashboardDataModeLive,
  DashboardPanelType,
  DashboardSearch,
} from "../types";
import { noop } from "./func";
import { useStorybookTheme } from "../hooks/useStorybookTheme";

type PanelStoryDecoratorProps = {
  definition: any;
  panelType: DashboardPanelType;
  panels?: {
    [key: string]: any;
  };
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
  panels = {},
  panelType,
  additionalProperties = {},
}: PanelStoryDecoratorProps) => {
  const { theme, wrapperRef } = useStorybookTheme();
  const { properties, ...rest } = definition;

  const newPanel = {
    name: `${panelType}.story`,
    panel_type: panelType,
    ...rest,
    properties: {
      ...(properties || {}),
      ...additionalProperties,
    },
    sql: "storybook",
  };

  return (
    <DashboardContext.Provider
      value={{
        versionMismatchCheck: false,
        ignoreEvents: false,
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
        dataMode: DashboardDataModeLive,
        snapshotId: null,
        dispatch: noop,
        error: null,
        dashboards: [],
        dashboardsMap: {},
        selectedPanel: null,
        selectedDashboard: {
          title: "Storybook Dashboard Wrapper",
          full_name: "storybook.dashboard.storybook_dashboard_wrapper",
          short_name: "storybook_dashboard_wrapper",
          type: "dashboard",
          tags: {},
          mod_full_name: "mod.storybook",
          is_top_level: true,
        },
        selectedDashboardInputs: {},
        lastChangedInput: null,
        panelsMap: {
          [newPanel.name]: newPanel,
          ...panels,
        },
        dashboard: {
          artificial: false,
          name: "storybook.dashboard.storybook_dashboard_wrapper",
          children: [newPanel],
          panel_type: "dashboard",
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
          theme,
          setTheme: noop,
          wrapperRef,
        },

        components: buildComponentsMap(),
        refetchDashboard: false,
        state: "complete",
        progress: 100,
        render: { headless: false, snapshotCompleteDiv: false },
        snapshot: null,
        snapshotFileName: null,
      }}
    >
      <Dashboard showPanelControls={false} />
    </DashboardContext.Provider>
  );
};
