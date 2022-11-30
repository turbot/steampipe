import { LeafNodeData, Width } from "../components/dashboards/common";
import { Ref } from "react";
import { Theme } from "../hooks/useTheme";

export interface IDashboardContext {
  versionMismatchCheck: boolean;
  ignoreEvents: boolean;
  metadata: DashboardMetadata | null;
  availableDashboardsLoaded: boolean;

  closePanelDetail(): void;
  dispatch(action: DashboardAction): void;

  dataMode: DashboardDataMode;
  snapshotId: string | null;

  refetchDashboard: boolean;

  error: any;

  panelsMap: PanelsMap;

  dashboards: AvailableDashboard[];
  dashboardsMap: AvailableDashboardsDictionary;
  dashboard: DashboardDefinition | null;

  selectedPanel: PanelDefinition | null;
  selectedDashboard: AvailableDashboard | null;
  selectedDashboardInputs: DashboardInputs;
  lastChangedInput: string | null;

  sqlDataMap: SQLDataMap;

  dashboardTags: DashboardTags;

  search: DashboardSearch;

  breakpointContext: IBreakpointContext;
  themeContext: IThemeContext;

  components: ComponentsMap;

  progress: number;
  state: DashboardRunState;
  render: {
    headless: boolean;
    snapshotCompleteDiv: boolean;
  };

  snapshot: DashboardSnapshot | null;
  snapshotFileName: string | null;
}

type DashboardSnapshotSchemaVersion = "20220614" | "20220929";

export interface DashboardSnapshot {
  schema_version: DashboardSnapshotSchemaVersion;
  layout: DashboardLayoutNode;
  panels: PanelsMap;
  inputs: DashboardInputs;
  variables: DashboardVariables;
  search_path: string[];
  start_time: string;
  end_time: string;
}

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

export const DashboardDataModeLive = "live";
export const DashboardDataModeCLISnapshot = "cli_snapshot";
export const DashboardDataModeCloudSnapshot = "cloud_snapshot";

export type DashboardDataMode = "live" | "cli_snapshot" | "cloud_snapshot";

export type SocketURLFactory = () => Promise<string>;

export interface IActions {
  [type: string]: string;
}

export interface ReceivedSocketMessagePayload {
  action: string;
  [key: string]: any;
}

export interface ComponentsMap {
  [name: string]: any;
}

export interface PanelsMap {
  [name: string]: PanelDefinition;
}

export type DashboardRunState = "ready" | "error" | "complete";

export const DashboardActions: IActions = {
  AVAILABLE_DASHBOARDS: "available_dashboards",
  CLEAR_DASHBOARD_INPUTS: "clear_dashboard_inputs",
  CONTROL_COMPLETE: "control_complete",
  CONTROL_ERROR: "control_error",
  CONTROLS_UPDATED: "controls_updated",
  DASHBOARD_METADATA: "dashboard_metadata",
  DELETE_DASHBOARD_INPUT: "delete_dashboard_input",
  EXECUTION_COMPLETE: "execution_complete",
  EXECUTION_ERROR: "execution_error",
  EXECUTION_STARTED: "execution_started",
  INPUT_VALUES_CLEARED: "input_values_cleared",
  LEAF_NODE_COMPLETE: "leaf_node_complete",
  LEAF_NODES_COMPLETE: "leaf_nodes_complete",
  LEAF_NODE_PROGRESS: "leaf_node_progress",
  SELECT_DASHBOARD: "select_dashboard",
  SELECT_PANEL: "select_panel",
  SET_DASHBOARD: "set_dashboard",
  SET_DASHBOARD_INPUT: "set_dashboard_input",
  SET_DASHBOARD_INPUTS: "set_dashboard_inputs",
  SET_DASHBOARD_SEARCH_VALUE: "set_dashboard_search_value",
  SET_DASHBOARD_SEARCH_GROUP_BY: "set_dashboard_search_group_by",
  SET_DASHBOARD_TAG_KEYS: "set_dashboard_tag_keys",
  SET_DATA_MODE: "set_data_mode",
  SET_REFETCH_DASHBOARD: "set_refetch_dashboard",
  WORKSPACE_ERROR: "workspace_error",
};

type DashboardExecutionEventSchemaVersion = "20220614" | "20220929";

export interface DashboardExecutionEventWithSchema {
  schema_version: DashboardExecutionEventSchemaVersion;
  [key: string]: any;
}

export interface DashboardExecutionCompleteEvent {
  action: string;
  schema_version: DashboardExecutionEventSchemaVersion;
  execution_id: string;
  snapshot: DashboardSnapshot;
}

// https://github.com/microsoft/TypeScript/issues/28046
export type ElementType<T extends ReadonlyArray<unknown>> =
  T extends ReadonlyArray<infer ElementType> ? ElementType : never;

const dashboardActions = Object.values(DashboardActions);

export type DashboardActionType = ElementType<typeof dashboardActions>;

export interface DashboardAction {
  type: DashboardActionType;
  [key: string]: any;
}

type DashboardSearchGroupByMode = "mod" | "tag";

interface DashboardSearchGroupBy {
  value: DashboardSearchGroupByMode;
  tag: string | null;
}

export interface DashboardSearch {
  value: string;
  groupBy: DashboardSearchGroupBy;
}

export interface DashboardTags {
  keys: string[];
}

export interface SelectedDashboardStates {
  dashboard_name: string | undefined;
  dataMode: DashboardDataMode;
  refetchDashboard: boolean;
  search: DashboardSearch;
  searchParams: URLSearchParams;
  selectedDashboard: AvailableDashboard | null;
  selectedDashboardInputs: DashboardInputs;
}

interface DashboardInputs {
  [name: string]: string;
}

interface DashboardVariables {
  [name: string]: any;
}
export interface ModDashboardMetadata {
  title: string;
  full_name: string;
  short_name: string;
}

interface InstalledModsDashboardMetadata {
  [key: string]: ModDashboardMetadata;
}

interface CliDashboardMetadata {
  version: string;
}

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

interface CloudDashboardMetadata {
  actor: CloudDashboardActorMetadata;
  identity: CloudDashboardIdentityMetadata;
  workspace: CloudDashboardWorkspaceMetadata;
}

export interface DashboardMetadata {
  mod: ModDashboardMetadata;
  installed_mods?: InstalledModsDashboardMetadata;
  cli?: CliDashboardMetadata;
  cloud?: CloudDashboardMetadata;
  telemetry: "info" | "none";
}

export type DashboardLayoutNode = {
  name: string;
  panel_type: DashboardPanelType;
  children?: DashboardLayoutNode[];
};

export type DashboardPanelType =
  | "benchmark"
  | "benchmark_tree"
  | "card"
  | "chart"
  | "container"
  | "control"
  | "dashboard"
  | "error"
  | "flow"
  | "graph"
  | "hierarchy"
  | "image"
  | "input"
  | "table"
  | "text";

export interface DashboardSnapshot {
  start_time: string;
  end_time: string;
  schema_version: DashboardSnapshotSchemaVersion;
  search_path: string[];
  layout: DashboardLayoutNode;
  panels: PanelsMap;
  variables: DashboardVariables;
  inputs: DashboardInputs;
}

interface AvailableDashboardTags {
  [key: string]: string;
}

type AvailableDashboardType = "benchmark" | "dashboard" | "snapshot";

export interface AvailableDashboard {
  full_name: string;
  short_name: string;
  mod_full_name?: string;
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

export interface ContainerDefinition {
  name: string;
  panel_type?: string;
  data?: LeafNodeData;
  title?: string;
  width?: number;
  children?: (ContainerDefinition | PanelDefinition)[];
}

export interface PanelProperties {
  [key: string]: any;
}

export interface SQLDataMap {
  [sql: string]: LeafNodeData;
}

export interface PanelDefinition {
  name: string;
  display?: string;
  display_type?: string;
  panel_type: DashboardPanelType;
  title?: string;
  description?: string;
  documentation?: string;
  width?: Width;
  sql?: string;
  data?: LeafNodeData;
  source_definition?: string;
  status?: DashboardRunState;
  error?: string;
  properties?: PanelProperties;
  dashboard: string;
  children?: DashboardLayoutNode[];
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

export interface DashboardsCollection {
  dashboards: AvailableDashboard[];
  dashboardsMap: AvailableDashboardsDictionary;
}

export interface DashboardDataOptions {
  dataMode: DashboardDataMode;
  snapshotId?: string;
}

export interface DashboardRenderOptions {
  headless?: boolean;
}
