import Error from "../../Error";
import Placeholder from "../../Placeholder";
import { BaseChartProps } from "../../charts";
import { CardProps } from "../../Card";
import { CheckProps } from "../../check/common";
import { classNames } from "../../../../utils/styles";
import { get } from "lodash";
import { getResponsivePanelWidthClass } from "../../../../utils/layout";
import { HierarchyProps } from "../../hierarchies";
import { ImageProps } from "../../Image";
import { InputProps } from "../../inputs";
import { memo, useState } from "react";
import {
  DashboardActions,
  PanelDefinition,
  useDashboard,
} from "../../../../hooks/useDashboard";
import { PanelProvider } from "../../../../hooks/usePanel";
import { TableProps } from "../../Table";
import { ThemeNames } from "../../../../hooks/useTheme";
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
  allowExpand?: boolean;
  forceBackground?: boolean;
  ready?: boolean;
  withTitle?: boolean;
}

const Panel = ({
  children,
  definition,
  allowExpand = true,
  forceBackground = false,
  ready = true,
  withTitle = true,
}: PanelProps) => {
  const [showZoomIcon, setShowZoomIcon] = useState(false);
  const [zoomIconClassName, setZoomIconClassName] =
    useState("text-black-scale-4");
  const {
    dispatch,
    themeContext: { theme },
  } = useDashboard();

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
        <section
          aria-labelledby={
            withTitle && definition.title
              ? `${definition.name}-title`
              : undefined
          }
          className={classNames(
            "col-span-12 m-0.5",
            forceBackground ||
              (definition.node_type !== "image" &&
                definition.node_type !== "card" &&
                definition.node_type !== "input") ||
              ((definition.node_type === "image" ||
                definition.node_type === "card" ||
                definition.node_type === "input") &&
                get(definition, "properties.type") === "table")
              ? "bg-dashboard-panel shadow-sm rounded-md"
              : null
          )}
        >
          {showZoomIcon && (
            <div
              className={classNames(
                "absolute cursor-pointer z-50 right-1 top-1",
                zoomIconClassName
              )}
              onClick={(e) => {
                e.stopPropagation();
                dispatch({
                  type: DashboardActions.SELECT_PANEL,
                  panel: { ...definition },
                });
              }}
            >
              <ZoomIcon className="h-5 w-5" />
            </div>
          )}
          {withTitle && definition.title && (
            <div
              className={classNames(
                definition.node_type === "input" &&
                  get(definition, "properties.type") !== "table"
                  ? "pl-0 pr-2 sm:pr-4 py-2"
                  : "px-4 py-4"
              )}
            >
              <h3
                id={`${definition.name}-title`}
                className="truncate"
                title={definition.title}
              >
                {definition.title}
              </h3>
            </div>
          )}

          <div
            className={classNames(
              withTitle &&
                definition.title &&
                ((definition.node_type !== "input" &&
                  definition.node_type !== "table") ||
                  (definition.node_type === "table" &&
                    get(definition, "properties.type") === "line"))
                ? classNames(
                    "border-t",
                    theme.name === ThemeNames.STEAMPIPE_DARK
                      ? "border-table-divide"
                      : "border-background"
                  )
                : null,
              (definition.node_type === "table" &&
                get(definition, "properties.type") !== "line") ||
                get(definition, "properties.type") === "table"
                ? "overflow-x-auto"
                : "overflow-x-hidden"
            )}
          >
            <PlaceholderComponent
              animate={!!children}
              ready={ready || !!definition.error}
            >
              <ErrorComponent error={definition.error} />
              <>{!definition.error ? children : null}</>
            </PlaceholderComponent>
          </div>
        </section>
      </div>
    </PanelProvider>
  );
};

export default memo(Panel);
