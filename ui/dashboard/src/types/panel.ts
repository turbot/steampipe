import { LeafNodeData, Width } from "../components/dashboards/common";
import { DashboardRunState } from "./dashboard";

export interface PanelProperties {
  [key: string]: any;
}

export interface PanelDefinition {
  name: string;
  display_type?: string;
  panel_type?: string;
  title?: string;
  description?: string;
  width?: Width;
  sql?: string;
  data?: LeafNodeData;
  source_definition?: string;
  status?: DashboardRunState;
  error?: Error;
  properties?: PanelProperties;
  dashboard: string;
}

export interface ContainerDefinition {
  name: string;
  panel_type?: string;
  allow_child_panel_expand?: boolean;
  data?: LeafNodeData;
  title?: string;
  width?: number;
  children?: (ContainerDefinition | PanelDefinition)[];
}
