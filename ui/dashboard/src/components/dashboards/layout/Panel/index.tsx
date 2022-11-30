import Error from "../../Error";
import PanelControls from "./PanelControls";
import PanelInformation from "./PanelInformation";
import PanelProgress from "./PanelProgress";
import PanelTitle from "../../titles/PanelTitle";
import Placeholder from "../../Placeholder";
import { BaseChartProps } from "../../charts/types";
import { CardProps } from "../../Card";
import { classNames } from "../../../../utils/styles";
import { getResponsivePanelWidthClass } from "../../../../utils/layout";
import { HierarchyProps } from "../../hierarchies/types";
import { ImageProps } from "../../Image";
import { InputProps } from "../../inputs/types";
import { memo, useState } from "react";
import { PanelDefinition } from "../../../../types";
import { PanelProvider, usePanel } from "../../../../hooks/usePanel";
import { ReactNode } from "react";
import { registerComponent } from "../../index";
import { TableProps } from "../../Table";
import { TextProps } from "../../Text";
import { useDashboard } from "../../../../hooks/useDashboard";

interface PanelProps {
  children: ReactNode;
  className?: string;
  definition:
    | BaseChartProps
    | CardProps
    | HierarchyProps
    | ImageProps
    | InputProps
    | PanelDefinition
    | TableProps
    | TextProps;
  showControls?: boolean;
  showPanelError?: boolean;
  forceBackground?: boolean;
  ready?: boolean;
}

const Panel = ({
  children,
  className,
  definition,
  showControls = true,
  showPanelError = true,
  forceBackground = false,
  ready = true,
}: PanelProps) => {
  const { selectedPanel } = useDashboard();
  const { panelControls, showPanelControls, setShowPanelControls } = usePanel();
  const [referenceElement, setReferenceElement] = useState(null);

  const baseStyles = classNames(
    "relative col-span-12",
    getResponsivePanelWidthClass(definition.width),
    "overflow-auto"
  );

  const ErrorComponent = Error;
  const PlaceholderComponent = Placeholder.component;
  const showPanelContents =
    !definition.error || (definition.error && !showPanelError);

  return (
    <div
      // @ts-ignore
      ref={setReferenceElement}
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
          <PanelControls
            referenceElement={referenceElement}
            controls={panelControls}
          />
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
            <PanelTitle name={definition.name} title={definition.title} />
          </div>
        )}

        <div
          className={classNames(
            "relative",
            definition.title &&
              ((definition.panel_type !== "input" &&
                definition.panel_type !== "table") ||
                (definition.panel_type === "table" &&
                  definition.display_type === "line"))
              ? "border-t border-divide"
              : null,
            selectedPanel ||
              (definition.panel_type === "table" &&
                definition.display_type !== "line") ||
              definition.display_type === "table"
              ? "overflow-x-auto"
              : "overflow-x-hidden",
            className
          )}
        >
          <PanelProgress className={definition.title ? null : "rounded-t-md"} />
          {showPanelContents && <PanelInformation />}
          <PlaceholderComponent
            animate={!!children}
            ready={ready || !!definition.error}
          >
            <ErrorComponent
              className={definition.title ? "rounded-t-none" : null}
              error={showPanelError && definition.error}
            />
            <>{showPanelContents ? children : null}</>
          </PlaceholderComponent>
        </div>
      </section>
    </div>
  );
};

const PanelWrapper = memo((props: PanelProps) => {
  const { children, ...rest } = props;
  return (
    <PanelProvider
      definition={props.definition}
      showControls={props.showControls}
    >
      <Panel {...rest}>{children}</Panel>
    </PanelProvider>
  );
});

registerComponent("panel", PanelWrapper);

export default PanelWrapper;
