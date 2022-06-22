import { ElementType } from "./common";
import { SocketURLFactory } from "./webSocket";

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
  DELETE_DASHBOARD_INPUT: "delete_dashboard_input",
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
  SET_DASHBOARD_INPUT: "set_dashboard_input",
  SET_DASHBOARD_INPUTS: "set_dashboard_inputs",
  SET_DASHBOARD_SEARCH_VALUE: "set_dashboard_search_value",
  SET_DASHBOARD_SEARCH_GROUP_BY: "set_dashboard_search_group_by",
  SET_DASHBOARD_TAG_KEYS: "set_dashboard_tag_keys",
  SET_DATA_MODE: "set_data_mode",
  SET_REFETCH_DASHBOARD: "set_refetch_dashboard",
  SET_SNAPSHOT: "set_snapshot",
  WORKSPACE_ERROR: "workspace_error",
};

const dashboardActions = Object.values(DashboardActions);

export type DashboardActionType = ElementType<typeof dashboardActions>;

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

/***** Props *****/

export interface DashboardProviderProps {
  children: null | JSX.Element | JSX.Element[];
  socketUrlFactory?: SocketURLFactory;
}

export interface EventHooks {
  [type: DashboardActionType]: (event: any) => Promise<void>;
}
