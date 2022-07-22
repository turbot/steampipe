import { classNames } from "../../../../utils/styles";
import {
  ContainerDefinition,
  DashboardDefinition,
} from "../../../../hooks/useDashboard";
import { getResponsivePanelWidthClass } from "../../../../utils/layout";

interface EventMap {
  [name: string]: any;
}

interface LayoutPanelProps {
  allowExpand?: boolean;
  children: null | JSX.Element | JSX.Element[];
  className?: string;
  definition: DashboardDefinition | ContainerDefinition;
  events?: EventMap;
  isDashboard?: boolean;
  withNarrowVertical?: boolean;
  withPadding?: boolean;
  withTitle?: boolean;
}

const LayoutPanel = ({
  children,
  className,
  definition,
  events = {},
  isDashboard = false,
  withNarrowVertical = false,
  withPadding = false,
  withTitle = true,
}: LayoutPanelProps) => {
  const panelWidthClass = getResponsivePanelWidthClass(definition.width);
  return (
    <div
      id={definition.name}
      className={classNames(
        "grid grid-cols-12 col-span-12 gap-x-4",
        withNarrowVertical ? "gap-y-2" : "gap-y-4 md:gap-y-6",
        panelWidthClass,
        withPadding ? "p-4" : null,
        "auto-rows-min",
        className
      )}
      {...events}
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
