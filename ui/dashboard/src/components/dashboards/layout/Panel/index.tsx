import PanelStatus from "./PanelStatus";
import PanelControls from "./PanelControls";
import PanelInformation from "./PanelInformation";
import PanelProgress from "./PanelProgress";
import PanelTitle from "../../titles/PanelTitle";
import Placeholder from "../../Placeholder";
import { BaseChartProps } from "../../charts/types";
import { CardProps } from "../../Card";
import { classNames } from "../../../../utils/styles";
import { DashboardPanelType, PanelDefinition } from "../../../../types";
import { FlowProps } from "../../flows/types";
import { getResponsivePanelWidthClass } from "../../../../utils/layout";
import { GraphProps } from "../../graphs/types";
import { HierarchyProps } from "../../hierarchies/types";
import { ImageProps } from "../../Image";
import { InputProps } from "../../inputs/types";
import { memo, useState } from "react";
import { PanelProvider, usePanel } from "../../../../hooks/usePanel";
import { ReactNode } from "react";
import { registerComponent } from "../../index";
import { TableProps } from "../../Table";
import { TextProps } from "../../Text";
import { useDashboard } from "../../../../hooks/useDashboard";

type PanelProps = {
  children: ReactNode;
  className?: string;
  definition:
    | BaseChartProps
    | CardProps
    | FlowProps
    | GraphProps
    | HierarchyProps
    | ImageProps
    | InputProps
    | PanelDefinition
    | TableProps
    | TextProps;
  parentType: DashboardPanelType;
  showControls?: boolean;
  showPanelError?: boolean;
  showPanelStatus?: boolean;
  forceBackground?: boolean;
};

const Panel = ({
  children,
  className,
  definition,
  showControls = true,
  showPanelError = true,
  showPanelStatus = true,
  forceBackground = false,
}: PanelProps) => {
  const { selectedPanel } = useDashboard();
  const {
    inputPanelsAwaitingValue,
    panelControls,
    showPanelControls,
    setShowPanelControls,
  } = usePanel();
  const [referenceElement, setReferenceElement] = useState(null);
  const baseStyles = classNames(
    "relative col-span-12",
    getResponsivePanelWidthClass(definition.width),
    "overflow-auto"
  );

  if (inputPanelsAwaitingValue.length > 0) {
    return null;
  }

  const PlaceholderComponent = Placeholder.component;

  const shouldShowLoader =
    showPanelStatus &&
    definition.status !== "cancelled" &&
    definition.status !== "error" &&
    definition.status !== "complete";

  const showPanelContents =
    !definition.error || (definition.error && !showPanelError);

  return (
    <div
      // @ts-ignore
      ref={setReferenceElement}
      id={definition.name}
      className={baseStyles}
      onMouseEnter={showControls ? () => setShowPanelControls(true) : undefined}
      onMouseLeave={() => setShowPanelControls(false)}
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
                definition.display_type !== "table" &&
                !forceBackground
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
                // @ts-ignore
                definition.status !== "complete") ||
                (definition.panel_type !== "input" &&
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
          <PanelInformation />
          <PlaceholderComponent
            animate={definition.status === "running"}
            ready={!shouldShowLoader}
          >
            <>
              {((showPanelError && definition.status === "error") ||
                showPanelStatus) && (
                <PanelStatus
                  definition={definition as PanelDefinition}
                  showPanelError={showPanelError}
                />
              )}
              {showPanelContents && children}
            </>
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
      parentType={props.parentType}
      showControls={props.showControls}
    >
      <Panel {...rest}>{children}</Panel>
    </PanelProvider>
  );
});

registerComponent("panel", PanelWrapper);

export default PanelWrapper;
