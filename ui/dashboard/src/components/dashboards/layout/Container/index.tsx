import Children from "../common/Children";
import LayoutPanel from "../common/LayoutPanel";
import { classNames } from "../../../../utils/styles";
import {
  ContainerDefinition,
  DashboardActions,
  useDashboard,
} from "../../../../hooks/useDashboard";
import { registerComponent } from "../../index";
import { useState } from "react";
import { ZoomIcon } from "../../../../constants/icons";

interface ContainerProps {
  allowChildPanelExpand?: boolean;
  allowExpand?: boolean;
  layoutDefinition?: ContainerDefinition;
  definition?: ContainerDefinition;
  expandDefinition: ContainerDefinition;
  withNarrowVertical?: boolean;
  withTitle?: boolean;
}

const showContainerForView = (
  container: ContainerDefinition,
  currentView: string
): boolean => {
  const containerViews = container.view || [];
  // If the container does not specify views and the current view is dashboard, show this container
  if (containerViews.length === 0 && currentView === "dashboard") {
    return true;
  }
  // If the container does not specify views and the current view is not dashboard, do not this container
  else if (containerViews.length === 0 && currentView !== "dashboard") {
    return false;
  }
  // Else see if the container views include the current view
  else {
    return containerViews.includes(currentView);
  }
};

const Container = ({
  allowChildPanelExpand = true,
  allowExpand = false,
  definition,
  expandDefinition,
  layoutDefinition,
  withNarrowVertical,
  withTitle,
}: ContainerProps) => {
  const [showZoomIcon, setShowZoomIcon] = useState(false);
  const { dispatch, panelsMap, view } = useDashboard();

  if (!definition && !layoutDefinition) {
    return null;
  }

  const panelDefinition = definition
    ? definition
    : layoutDefinition && panelsMap[layoutDefinition.name]
    ? panelsMap[layoutDefinition.name]
    : layoutDefinition;

  // Check if this panel should be shown according to the current view
  if (
    !panelDefinition ||
    !showContainerForView(panelDefinition as ContainerDefinition, view)
  ) {
    return null;
  }

  return (
    <LayoutPanel
      allowExpand={allowExpand}
      className="relative"
      definition={panelDefinition}
      events={{
        onMouseEnter: allowExpand
          ? () => {
              setShowZoomIcon(true);
            }
          : undefined,

        onMouseLeave: allowExpand
          ? () => {
              setShowZoomIcon(false);
            }
          : undefined,
      }}
      withNarrowVertical={withNarrowVertical}
      withTitle={withTitle}
    >
      <>
        {showZoomIcon && (
          <div
            className={classNames(
              "absolute cursor-pointer z-50 right-1 top-1 text-black-scale-4"
            )}
            onClick={(e) => {
              e.stopPropagation();
              dispatch({
                type: DashboardActions.SELECT_PANEL,
                panel: { ...expandDefinition },
                // panel: {
                //   ...{
                //     ...panelDefinition,
                //     children: definition
                //       ? definition.children
                //       : layoutDefinition
                //       ? layoutDefinition.children
                //       : [],
                //   },
                // },
              });
            }}
          >
            <ZoomIcon className="h-5 w-5" />
          </div>
        )}
      </>
      <Children
        allowPanelExpand={allowChildPanelExpand}
        children={
          definition
            ? definition.children
            : layoutDefinition
            ? layoutDefinition.children
            : []
        }
      />
    </LayoutPanel>
  );
};

registerComponent("container", Container);

export default Container;
