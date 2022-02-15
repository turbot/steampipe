import { classNames } from "../../../../utils/styles";
import {
  ContainerDefinition,
  DashboardDefinition,
} from "../../../../hooks/useDashboard";
import { getResponsivePanelWidthClass } from "../../../../utils/layout";

interface LayoutPanelProps {
  children: null | JSX.Element | JSX.Element[];
  definition: DashboardDefinition | ContainerDefinition;
  withPadding?: boolean;
  withNarrowVertical?: boolean;
}

const LayoutPanel = ({
  children,
  definition,
  withPadding = false,
  withNarrowVertical = false,
}: LayoutPanelProps) => (
  <div
    className={classNames(
      "grid grid-cols-12 gap-x-4",
      withNarrowVertical ? "gap-y-2" : "gap-y-6",
      "col-span-12",
      getResponsivePanelWidthClass(definition.width),
      withPadding ? "p-4" : null,
      "auto-rows-min"
    )}
  >
    {children}
  </div>
);

export default LayoutPanel;
