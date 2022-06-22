import { IActions } from "./dashboard";

/***** Actions *****/

export type SocketURLFactory = () => Promise<string>;

export const SocketActions: IActions = {
  CLEAR_DASHBOARD: "clear_dashboard",
  GET_AVAILABLE_DASHBOARDS: "get_available_dashboards",
  GET_DASHBOARD_METADATA: "get_dashboard_metadata",
  SELECT_DASHBOARD: "select_dashboard",
  INPUT_CHANGED: "input_changed",
};

/***** Messages *****/

export interface ReceivedSocketMessagePayload {
  action: string;
  [key: string]: any;
}
