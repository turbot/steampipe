import { classNames } from "../../../../utils/styles";
import {
  ContainerDefinition,
  DashboardDefinition,
} from "../../../../hooks/useDashboard";
import { getResponsivePanelWidthClass } from "../../../../utils/layout";

interface LayoutPanelProps {
  children: null | JSX.Element | JSX.Element[];
  definition: DashboardDefinition | ContainerDefinition;
  isDashboard?: boolean;
  withPadding?: boolean;
  withTitle?: boolean;
}

const LayoutPanel = ({
  children,
  definition,
  isDashboard = false,
  withPadding = false,
  withTitle = true,
}: LayoutPanelProps) => {
  const panelWidthClass = getResponsivePanelWidthClass(definition.width);
  return (
    <div
      className={classNames(
        "grid grid-cols-12 gap-x-4 gap-y-4 md:gap-y-6 col-span-12",
        panelWidthClass,
        withPadding ? "p-4" : null,
        "auto-rows-min"
      )}
    >
      {withTitle && definition.title && isDashboard && (
        <h1 className={classNames("col-span-12")}>{definition.title}</h1>
      )}
      {withTitle && definition.title && !isDashboard && (
        <h2 className={classNames("col-span-12")}>{definition.title}</h2>
      )}
      {children}
    </div>
  );
};

export default LayoutPanel;
