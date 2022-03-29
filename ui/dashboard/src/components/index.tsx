import ExternalLink from "./ExternalLink";
import { ComponentsMap } from "../hooks/useDashboard";

const buildComponentsMap = (overrides = {}): ComponentsMap => {
  return {
    ExternalLink,
    ...overrides,
  };
};

export { buildComponentsMap };
