import { classNames } from "../../../../utils/styles";
import {
  ContainerDefinition,
  ReportDefinition,
} from "../../../../hooks/useReport";
import { getResponsivePanelWidthClass } from "../../../../utils/layout";
import { has } from "lodash";

interface LayoutPanelProps {
  children: null | JSX.Element | JSX.Element[];
  definition: ReportDefinition | ContainerDefinition;
  withPadding?: boolean;
}

const LayoutPanel = ({
  children,
  definition,
  withPadding = false,
}: LayoutPanelProps) => (
  <div
    className={classNames(
      "grid grid-cols-12 gap-4 col-span-12",
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
