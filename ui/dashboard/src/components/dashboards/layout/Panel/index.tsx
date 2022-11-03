import Error from "../../Error";
import PanelProgress from "./PanelProgress";
import Placeholder from "../../Placeholder";
import { BaseChartProps } from "../../charts/types";
import {
  BenchmarkDefinition,
  DashboardActions,
  PanelDefinition,
} from "../../../../types";
import { CardProps } from "../../Card";
import { classNames } from "../../../../utils/styles";
import { getResponsivePanelWidthClass } from "../../../../utils/layout";
import { HierarchyProps } from "../../hierarchies/types";
import { ImageProps } from "../../Image";
import { InputProps } from "../../inputs/types";
import { memo, useState } from "react";
import { PanelProvider } from "../../../../hooks/usePanel";
import { ReactNode } from "react";
import { registerComponent } from "../../index";
import { TableProps } from "../../Table";
import { TextProps } from "../../Text";
import { ThemeNames } from "../../../../hooks/useTheme";
import { useDashboard } from "../../../../hooks/useDashboard";
import { ZoomIcon } from "../../../../constants/icons";

interface PanelProps {
  children: ReactNode;
  className?: string;
  definition:
    | BaseChartProps
    | BenchmarkDefinition
    | CardProps
    | HierarchyProps
    | ImageProps
    | InputProps
    | PanelDefinition
    | TableProps
    | TextProps;
  showControls?: boolean;
  forceBackground?: boolean;
  ready?: boolean;
}

const Panel = memo(
  ({
    children,
    className,
    definition,
    showControls = true,
    forceBackground = false,
    ready = true,
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
        showControls={showControls}
        setZoomIconClassName={setZoomIconClassName}
      >
        <div
          id={definition.name}
          className={baseStyles}
          onMouseEnter={
            showControls
              ? () => {
                  setShowZoomIcon(true);
                }
              : undefined
          }
          onMouseLeave={
            showControls
              ? () => {
                  setShowZoomIcon(false);
                }
              : undefined
          }
        >
          <section
            aria-labelledby={
              definition.title ? `${definition.name}-title` : undefined
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
                    panel: definition,
                  });
                }}
              >
                <ZoomIcon className="h-5 w-5" />
              </div>
            )}
            {definition.title && (
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
                (definition.panel_type === "table" &&
                  definition.display_type !== "line") ||
                  definition.display_type === "table"
                  ? "overflow-x-auto"
                  : "overflow-x-hidden",
                className
              )}
            >
              <PanelProgress
                className={definition.title ? null : "rounded-t-md"}
              />
              <PlaceholderComponent
                animate={!!children}
                ready={ready || !!definition.error}
              >
                <ErrorComponent
                  className={definition.title ? "rounded-t-none" : null}
                  error={definition.error}
                />
                <>{!definition.error ? children : null}</>
              </PlaceholderComponent>
            </div>
          </section>
        </div>
      </PanelProvider>
    );
  }
);

registerComponent("panel", Panel);

export default Panel;
