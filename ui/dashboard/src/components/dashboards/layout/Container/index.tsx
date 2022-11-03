import Children from "../Children";
import Grid from "../Grid";
import { classNames } from "../../../../utils/styles";
import { ContainerDefinition } from "../../../../types";
import { registerComponent } from "../../index";
import { useDashboard } from "../../../../hooks/useDashboard";

interface ContainerProps {
  layoutDefinition?: ContainerDefinition;
  definition?: ContainerDefinition;
  expandDefinition: ContainerDefinition;
  showChildPanelControls?: boolean;
  showControls?: boolean;
  withNarrowVertical?: boolean;
  withTitle?: boolean;
}

const Container = ({
  definition,
  layoutDefinition,
  showChildPanelControls = true,
  showControls = false,
}: ContainerProps) => {
  // const [showZoomIcon, setShowZoomIcon] = useState(false);
  const { panelsMap } = useDashboard();

  if (!definition && !layoutDefinition) {
    return null;
  }

  const panelDefinition = definition
    ? definition
    : layoutDefinition && panelsMap[layoutDefinition.name]
    ? panelsMap[layoutDefinition.name]
    : layoutDefinition;

  if (!panelDefinition) {
    return null;
  }

  const title = panelDefinition.title ? (
    <h2 className={classNames("col-span-12")}>{panelDefinition.title}</h2>
  ) : null;

  return (
    <Grid name={panelDefinition.name} width={panelDefinition.width}>
      {title}
      <Children
        children={
          definition
            ? definition.children
            : layoutDefinition
            ? layoutDefinition.children
            : []
        }
        showPanelControls={showChildPanelControls}
      />
    </Grid>
  );

  // return (
  //   <LayoutPanel
  //     showControls={showControls}
  //     className="relative"
  //     definition={panelDefinition}
  //     events={{
  //       onMouseEnter: showControls
  //         ? () => {
  //             setShowZoomIcon(true);
  //           }
  //         : undefined,
  //
  //       onMouseLeave: showControls
  //         ? () => {
  //             setShowZoomIcon(false);
  //           }
  //         : undefined,
  //     }}
  //     withNarrowVertical={withNarrowVertical}
  //     withTitle={withTitle}
  //   >
  //     <>
  //       {showZoomIcon && (
  //         <div
  //           className={classNames(
  //             "absolute cursor-pointer z-50 right-1 top-1 text-black-scale-4"
  //           )}
  //           onClick={(e) => {
  //             e.stopPropagation();
  //             dispatch({
  //               type: DashboardActions.SELECT_PANEL,
  //               panel: { ...expandDefinition },
  //               // panel: {
  //               //   ...{
  //               //     ...panelDefinition,
  //               //     children: definition
  //               //       ? definition.children
  //               //       : layoutDefinition
  //               //       ? layoutDefinition.children
  //               //       : [],
  //               //   },
  //               // },
  //             });
  //           }}
  //         >
  //           <ZoomIcon className="h-5 w-5" />
  //         </div>
  //       )}
  //     </>
  //     <Children
  //       showPanelControls={showChildPanelControls}
  //       children={
  //         definition
  //           ? definition.children
  //           : layoutDefinition
  //           ? layoutDefinition.children
  //           : []
  //       }
  //     />
  //   </LayoutPanel>
  // );
};

registerComponent("container", Container);

export default Container;
