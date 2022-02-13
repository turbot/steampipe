import { classNames } from "../../../../utils/styles";
import {
  ContainerDefinition,
  DashboardDefinition,
} from "../../../../hooks/useDashboard";
import { getResponsivePanelWidthClass } from "../../../../utils/layout";
import { has } from "lodash";

interface LayoutPanelProps {
  children: null | JSX.Element | JSX.Element[];
  definition: DashboardDefinition | ContainerDefinition;
  withPadding?: boolean;
}

const LayoutPanel = ({
  children,
  definition,
  withPadding = false,
}: LayoutPanelProps) => (
  <div
    className={classNames(
      "grid grid-cols-12 gap-x-4 gap-y-6 col-span-12 auto-rows-min",
      has(definition, "width")
        ? // @ts-ignore
          getResponsivePanelWidthClass(definition.width)
        : null,
      withPadding ? "p-4" : null
    )}
  >
    {children}
  </div>
);

export default LayoutPanel;
