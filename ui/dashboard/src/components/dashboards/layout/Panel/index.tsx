import Error from "../../Error";
import Icon from "../../../Icon";
import PanelProgress from "./PanelProgress";
import Placeholder from "../../Placeholder";
import useDownloadPanelData from "../../../../hooks/useDownloadPanelData";
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
import { memo, useCallback, useEffect, useState } from "react";
import { PanelProvider } from "../../../../hooks/usePanel";
import { ReactNode } from "react";
import { registerComponent } from "../../index";
import { TableProps } from "../../Table";
import { TextProps } from "../../Text";
import { ThemeNames } from "../../../../hooks/useTheme";
import { useDashboard } from "../../../../hooks/useDashboard";

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

const PanelControl = ({ action, icon, title }) => {
  return (
    <div
      className="p-1 cursor-pointer bg-black-scale-2 text-foreground last:rounded-tr-[4px]"
      onClick={async (e) => await action(e)}
      title={title}
    >
      <Icon className="w-5 h-5" icon={icon} />
    </div>
  );
};

const Panel = memo(
  ({
    children,
    className,
    definition,
    showControls = true,
    forceBackground = false,
    ready = true,
  }: PanelProps) => {
    const {
      dispatch,
      themeContext: { theme },
    } = useDashboard();
    const { download } = useDownloadPanelData(definition as PanelDefinition);

    const openPanelDetail = useCallback(
      (e) => {
        e.stopPropagation();
        dispatch({
          type: DashboardActions.SELECT_PANEL,
          panel: definition,
        });
      },
      [dispatch, definition]
    );

    const downloadPanelData = useCallback(
      async (e) => {
        e.stopPropagation();
        await download();
      },
      [dispatch, definition]
    );

    const defaultPanelControls = [
      {
        action: downloadPanelData,
        icon: "arrow-down-tray",
        title: "Download data",
      },
      {
        action: openPanelDetail,
        icon: "arrows-pointing-out",
        title: "View detail",
      },
    ];
    const [panelControls, setPanelControls] = useState(() =>
      showControls && definition && definition.data
        ? defaultPanelControls
        : showControls
        ? [defaultPanelControls[1]]
        : []
    );
    const [showPanelControls, setShowPanelControls] = useState(false);

    useEffect(() => {
      if (!definition || !definition.data) {
        return;
      }
      setPanelControls([...defaultPanelControls]);
    }, [definition]);

    const baseStyles = classNames(
      "relative col-span-12",
      getResponsivePanelWidthClass(definition.width),
      "overflow-auto"
    );

    const ErrorComponent = Error;
    const PlaceholderComponent = Placeholder.component;

    return (
      <PanelProvider definition={definition} showControls={showControls}>
        <div
          id={definition.name}
          className={baseStyles}
          onMouseEnter={
            showControls
              ? () => {
                  setShowPanelControls(true);
                }
              : undefined
          }
          onMouseLeave={
            showControls
              ? () => {
                  setShowPanelControls(false);
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
            {showPanelControls && (
              <div
                className={classNames(
                  "absolute drop-shadow-sm z-50 right-1 top-1"
                )}
              >
                <div className="flex space-x-px">
                  {panelControls.map((panelControl, idx) => (
                    <PanelControl
                      key={idx}
                      action={panelControl.action}
                      icon={panelControl.icon}
                      title={panelControl.title}
                    />
                  ))}
                </div>
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
