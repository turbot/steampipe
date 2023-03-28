import DashboardListEmptyCallToAction from "./DashboardListEmptyCallToAction";
import { ComponentsMap } from "../types";

const buildComponentsMap = (overrides = {}): ComponentsMap => {
  return {
    DashboardListEmptyCallToAction,
    ...overrides,
  };
};

export { buildComponentsMap };
