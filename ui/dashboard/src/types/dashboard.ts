import { ContainerDefinition, PanelDefinition } from "./panel";
import { ElementType } from "./common";
import { LeafNodeData } from "../components/dashboards/common";
import { Ref } from "react";
import { SocketURLFactory } from "./webSocket";
import { Theme } from "../hooks/useTheme";

/***** Actions *****/

export interface IActions {
  [type: string]: string;
}

export const DashboardActions: IActions = {
  AVAILABLE_DASHBOARDS: "available_dashboards",
  CLEAR_DASHBOARD_INPUTS: "clear_dashboard_inputs",
  CLEAR_SNAPSHOT: "clear_snapshot",
  CONTROL_COMPLETE: "control_complete",
  CONTROL_ERROR: "control_error",
  DASHBOARD_METADATA: "dashboard_metadata",
  EXECUTION_COMPLETE: "execution_complete",
  EXECUTION_ERROR: "execution_error",
  EXECUTION_STARTED: "execution_started",
  INPUT_VALUES_CLEARED: "input_values_cleared",
  LEAF_NODE_COMPLETE: "leaf_node_complete",
  LEAF_NODE_PROGRESS: "leaf_node_progress",
  SELECT_DASHBOARD: "select_dashboard",
  SELECT_PANEL: "select_panel",
  SELECT_SNAPSHOT: "select_snapshot",
  SET_DASHBOARD: "set_dashboard",
  SET_DASHBOARD_TAG_KEYS: "set_dashboard_tag_keys",
  SET_DATA_MODE: "set_data_mode",
  SET_REFETCH_DASHBOARD: "set_refetch_dashboard",
  SET_SNAPSHOT: "set_snapshot",
  WORKSPACE_ERROR: "workspace_error",
};

const dashboardActions = Object.values(DashboardActions);

export type DashboardActionType = ElementType<typeof dashboardActions>;

export interface DashboardAction {
  type: DashboardActionType;
  [key: string]: any;
}

/***** Dashboard Metadata *****/

export interface CloudDashboardActorMetadata {
  id: string;
  handle: string;
}

export interface CloudDashboardIdentityMetadata {
  id: string;
  handle: string;
  type: "org" | "user";
}

export interface CloudDashboardWorkspaceMetadata {
  id: string;
  handle: string;
}

export interface CloudDashboardMetadata {
  actor: CloudDashboardActorMetadata;
  identity: CloudDashboardIdentityMetadata;
  workspace: CloudDashboardWorkspaceMetadata;
}

export interface ModDashboardMetadata {
  title: string;
  full_name: string;
  short_name: string;
}

export interface InstalledModsDashboardMetadata {
  [key: string]: ModDashboardMetadata;
}

export interface DashboardMetadata {
  mod: ModDashboardMetadata;
  installed_mods?: InstalledModsDashboardMetadata;
  cloud?: CloudDashboardMetadata;
  telemetry: "info" | "none";
}

export interface AvailableDashboardTags {
  [key: string]: string;
}

export type AvailableDashboardType = "benchmark" | "dashboard";

export interface AvailableDashboard {
  full_name: string;
  short_name: string;
  mod_full_name: string;
  tags: AvailableDashboardTags;
  title: string;
  is_top_level: boolean;
  type: AvailableDashboardType;
  children?: AvailableDashboard[];
  trunks?: string[][];
}

export interface AvailableDashboardsDictionary {
  [key: string]: AvailableDashboard;
}

export interface DashboardsCollection {
  dashboards: AvailableDashboard[];
  dashboardsMap: AvailableDashboardsDictionary;
}

export interface DashboardTags {
  keys: string[];
}

/***** Mode *****/

export type DashboardDataMode = "live" | "snapshot";

/***** Search *****/

export type DashboardSearchGroupByMode = "mod" | "tag";

export interface DashboardSearchGroupBy {
  value: DashboardSearchGroupByMode;
  tag: string | null;
}

export interface DashboardSearch {
  value: string;
  groupBy: DashboardSearchGroupBy;
}

/***** Inputs *****/

export interface DashboardInputs {
  [name: string]: string;
}

/***** Hooks *****/

export interface EventHooks {
  [type: DashboardActionType]: (event: any) => Promise<void>;
}

/***** Props *****/

export interface DashboardProviderProps {
  analyticsContext: any;
  breakpointContext: any;
  children: null | JSX.Element | JSX.Element[];
  componentOverrides?: {};
  eventHooks?: EventHooks;
  featureFlags?: string[];
  socketUrlFactory?: SocketURLFactory;
  stateDefaults?: {};
  themeContext: any;
}

/***** Components *****/

export interface ComponentsMap {
  [name: string]: any;
}

/***** Context *****/

export interface IBreakpointContext {
  currentBreakpoint: string | null;
  maxBreakpoint(breakpointAndDown: string): boolean;
  minBreakpoint(breakpointAndUp: string): boolean;
  width: number;
}

export interface IThemeContext {
  theme: Theme;
  setTheme(theme: string): void;
  wrapperRef: Ref<null>;
}

/***** Dashboard State *****/

export interface SelectedDashboardStates {
  dashboard_name: string | null;
  selectedDashboard: AvailableDashboard | null;
}

export interface DashboardDefinition {
  artificial: boolean;
  name: string;
  panel_type: string;
  title?: string;
  width?: number;
  children?: (ContainerDefinition | PanelDefinition)[];
  dashboard: string;
}

export type DashboardRunState = "running" | "error" | "complete";

export interface PanelsMap {
  [name: string]: PanelDefinition;
}

export interface SQLDataMap {
  [sql: string]: LeafNodeData;
}
