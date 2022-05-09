import DashboardListEmptyCallToAction from "./DashboardListEmptyCallToAction";
import ExternalLink from "./ExternalLink";
import { ComponentsMap } from "../hooks/useDashboard";

const buildComponentsMap = (overrides = {}): ComponentsMap => {
  return {
    DashboardListEmptyCallToAction,
    ExternalLink,
    ...overrides,
  };
};

export { buildComponentsMap };
