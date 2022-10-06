import Error from "../../Error";
import Placeholder from "../../Placeholder";
import { BaseChartProps } from "../../charts/types";
import { BenchmarkDefinition, PanelDefinition } from "../../../../types";
import { CardProps } from "../../Card";
import { classNames } from "../../../../utils/styles";
import { DashboardActions, useDashboard } from "../../../../hooks/useDashboard";
import { getResponsivePanelWidthClass } from "../../../../utils/layout";
import { HierarchyProps } from "../../hierarchies/types";
import { ImageProps } from "../../Image";
import { InputProps } from "../../inputs/types";
import { memo, useState } from "react";
import { PanelProvider } from "../../../../hooks/usePanel";
import { registerComponent } from "../../index";
import { TableProps } from "../../Table";
import { ThemeNames } from "../../../../hooks/useTheme";
import { TextProps } from "../../Text";
import { ZoomIcon } from "../../../../constants/icons";

interface PanelProps {
  children: null | JSX.Element | JSX.Element[];
  definition:
    | BaseChartProps
    | CardProps
    | HierarchyProps
    | ImageProps
    | InputProps
    | PanelDefinition
    | TableProps
    | TextProps;
  layoutDefinition:
    | BaseChartProps
    | CardProps
    | HierarchyProps
    | ImageProps
    | InputProps
    | PanelDefinition
    | BenchmarkDefinition
    | TableProps
    | TextProps;
  allowExpand?: boolean;
  forceBackground?: boolean;
  ready?: boolean;
  withOverflow?: boolean;
  withTitle?: boolean;
}

interface PanelWrapperProps {
  children: (definition: PanelDefinition) => null | JSX.Element | JSX.Element[];
  layoutDefinition:
    | BaseChartProps
    | CardProps
    | HierarchyProps
    | ImageProps
    | InputProps
    | PanelDefinition
    | BenchmarkDefinition
    | TableProps
    | TextProps;
  allowExpand?: boolean;
  forceBackground?: boolean;
  ready?: (PanelDefinition) => boolean;
  withOverflow?: boolean;
  withTitle?: boolean;
}

const Panel = memo(
  ({
    children,
    definition,
    layoutDefinition,
    allowExpand = true,
    forceBackground = false,
    ready = true,
    withOverflow = false,
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
                (definition.panel_type !== "image" &&
                  definition.panel_type !== "card" &&
                  definition.panel_type !== "input") ||
                ((definition.panel_type === "image" ||
                  definition.panel_type === "card" ||
                  definition.panel_type === "input") &&
                  definition.display_type === "table")
                ? "bg-dashboard-panel print:bg-white shadow-sm rounded-md"
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
                    panel: { ...(layoutDefinition || {}), ...definition },
                  });
                }}
              >
                <ZoomIcon className="h-5 w-5" />
              </div>
            )}
            {withTitle && definition.title && (
              <div
                className={classNames(
                  definition.panel_type === "input" &&
                    definition.display_type !== "table"
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
                  ((definition.panel_type !== "input" &&
                    definition.panel_type !== "table") ||
                    (definition.panel_type === "table" &&
                      definition.display_type === "line"))
                  ? classNames(
                      "border-t",
                      theme.name === ThemeNames.STEAMPIPE_DARK
                        ? "border-table-divide"
                        : "border-background"
                    )
                  : null,
                withOverflow ||
                  (definition.panel_type === "table" &&
                    definition.display_type !== "line") ||
                  definition.display_type === "table"
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
  }
);

const PanelWrapper = ({
  children,
  allowExpand = true,
  forceBackground = false,
  layoutDefinition,
  ready = () => true,
  withOverflow = false,
  withTitle = true,
}: PanelWrapperProps) => {
  const { panelsMap } = useDashboard();
  const panel = panelsMap[layoutDefinition.name];
  return (
    <Panel
      allowExpand={allowExpand}
      definition={panel || layoutDefinition}
      layoutDefinition={layoutDefinition}
      forceBackground={forceBackground}
      ready={ready(panel || layoutDefinition)}
      withOverflow={withOverflow}
      withTitle={withTitle}
    >
      {children(panel || layoutDefinition)}
    </Panel>
  );
};

registerComponent("panel", PanelWrapper);

export default PanelWrapper;
