import DashboardListEmptyCallToAction from "./DashboardListEmptyCallToAction";
import ExternalLink from "./ExternalLink";
import { ComponentsMap } from "../types/dashboard";

const buildComponentsMap = (overrides = {}): ComponentsMap => {
  return {
    DashboardListEmptyCallToAction,
    ExternalLink,
    ...overrides,
  };
};

export { buildComponentsMap };
