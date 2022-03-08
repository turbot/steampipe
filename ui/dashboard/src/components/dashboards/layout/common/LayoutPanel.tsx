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
  withNarrowVertical?: boolean;
  withPadding?: boolean;
  withTitle?: boolean;
}

const LayoutPanel = ({
  children,
  definition,
  isDashboard = false,
  withNarrowVertical = false,
  withPadding = false,
  withTitle = true,
}: LayoutPanelProps) => {
  const panelWidthClass = getResponsivePanelWidthClass(definition.width);
  return (
    <div
      className={classNames(
        "grid grid-cols-12 gap-x-4",
        withNarrowVertical ? "gap-y-2" : "gap-y-6",
        "col-span-12",
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
      {/*<section*/}
      {/*  className="col-span-12"*/}
      {/*  aria-labelledby={*/}
      {/*    definition.title ? `${definition.name}-title` : undefined*/}
      {/*  }*/}
      {/*>*/}
      {children}
      {/*</section>*/}
    </div>
  );
};

export default LayoutPanel;
