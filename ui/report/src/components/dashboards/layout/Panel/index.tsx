import Error from "../../Error";
import Placeholder from "../../Placeholder";
import React, { memo } from "react";
import useDimensions from "../../../../hooks/useDimensions";
import { BaseChartProps } from "../../charts";
import { CardProps } from "../../Card";
import { classNames } from "../../../../utils/styles";
import { getResponsivePanelWidthClass } from "../../../../utils/layout";
import { PanelDefinition } from "../../../../hooks/useDashboard";
import { PanelProvider } from "../../../../hooks/usePanel";
import { TableProps } from "../../Table";

interface PanelProps {
  children: null | JSX.Element | JSX.Element[];
  definition:
    | BaseChartProps
    | CardProps
    // | ImageProps
    | PanelDefinition
    | TableProps;
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
    definition.width ? getResponsivePanelWidthClass(definition.width) : null,
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
      <div
        ref={panelRef}
        id={definition.name}
        className={baseStyles}
        // style={{ height: height ? `${height}px` : "auto" }}
      >
        <div className="col-span-12">
          <PlaceholderComponent
            animate={!!children}
            ready={ready || !!definition.error}
            // type={type}
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
