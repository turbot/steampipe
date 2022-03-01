import Error from "../../Error";
import Placeholder from "../../Placeholder";
import { BaseChartProps } from "../../charts";
import { CardProps } from "../../Card";
import { CheckProps } from "../../check/common";
import { classNames } from "../../../../utils/styles";
import { getResponsivePanelWidthClass } from "../../../../utils/layout";
import { HierarchyProps } from "../../hierarchies";
import { ImageProps } from "../../Image";
import { InputProps } from "../../inputs";
import { memo, useState } from "react";
import { PanelDefinition, useDashboard } from "../../../../hooks/useDashboard";
import { PanelProvider } from "../../../../hooks/usePanel";
import { TableProps } from "../../Table";
import { TextProps } from "../../Text";
import { ZoomIcon } from "../../../../constants/icons";

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
  allowExpand?: boolean;
}

const Panel = ({
  children,
  definition,
  allowExpand = true,
  ready = true,
}: PanelProps) => {
  const [showZoomIcon, setShowZoomIcon] = useState(false);
  const [zoomIconClassName, setZoomIconClassName] =
    useState("text-black-scale-4");
  const { dispatch } = useDashboard();

  const baseStyles = classNames(
    "relative col-span-12",
    getResponsivePanelWidthClass(definition.width),
    "overflow-auto"
  );

  const ErrorComponent = Error;
  const PlaceholderComponent = Placeholder.component;

  return (
    <PanelProvider
      definition={definition}
      allowExpand={allowExpand}
      setZoomIconClassName={setZoomIconClassName}
    >
      <div
        id={definition.name}
        className={baseStyles}
        onMouseEnter={
          allowExpand
            ? () => {
                setShowZoomIcon(true);
              }
            : undefined
        }
        onMouseLeave={
          allowExpand
            ? () => {
                setShowZoomIcon(false);
              }
            : undefined
        }
      >
        {showZoomIcon && (
          <div
            className={classNames(
              "absolute cursor-pointer z-50 right-1 top-1",
              zoomIconClassName
            )}
            onClick={() =>
              dispatch({
                type: "select_panel",
                panel: { ...definition },
              })
            }
          >
            <ZoomIcon className="h-5 w-5" />
          </div>
        )}
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
