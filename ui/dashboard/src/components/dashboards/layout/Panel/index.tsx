import Error from "../../Error";
import Placeholder from "../../Placeholder";
import useDimensions from "../../../../hooks/useDimensions";
import { BaseChartProps } from "../../charts";
import { CardProps } from "../../Card";
import { CheckProps } from "../../check/common";
import { classNames } from "../../../../utils/styles";
import { getResponsivePanelWidthClass } from "../../../../utils/layout";
import { HierarchyProps } from "../../hierarchies";
import { ImageProps } from "../../Image";
import { InputProps } from "../../inputs";
import { memo } from "react";
import { PanelDefinition } from "../../../../hooks/useDashboard";
import { PanelProvider } from "../../../../hooks/usePanel";
import { TableProps } from "../../Table";
import { TextProps } from "../../Text";

interface PanelProps {
  children: null | JSX.Element | JSX.Element[];
  definition:
    | BaseChartProps
    | CardProps
    | CheckProps
    | HierarchyProps
    | ImageProps
    | InputProps
    | PanelDefinition
    | TableProps
    | TextProps;
  ready?: boolean;
  showExpand?: boolean;
}

const Panel = ({
  children,
  definition,
  showExpand = true,
  ready = true,
}: PanelProps) => {
  const [panelRef, dimensions] = useDimensions();

  const baseStyles = classNames(
    "col-span-12",
    getResponsivePanelWidthClass(definition.width),
    "overflow-auto"
  );

  const ErrorComponent = Error;
  const PlaceholderComponent = Placeholder.component;

  return (
    <PanelProvider
      definition={definition}
      dimensions={dimensions}
      showExpand={showExpand}
    >
      <div ref={panelRef} id={definition.name} className={baseStyles}>
        <div className="col-span-12">
          <PlaceholderComponent
            animate={!!children}
            ready={ready || !!definition.error}
          >
            <ErrorComponent error={definition.error} />
            <>{!definition.error ? children : null}</>
          </PlaceholderComponent>
        </div>
      </div>
    </PanelProvider>
  );
};

export default memo(Panel);
