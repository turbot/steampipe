import Children from "../Children";
import Grid from "../Grid";
import { classNames } from "../../../../utils/styles";
import { ContainerDefinition } from "../../../../types";
import { registerComponent } from "../../index";
import { useDashboard } from "../../../../hooks/useDashboard";

interface ContainerProps {
  allowChildPanelExpand?: boolean;
  allowExpand?: boolean;
  layoutDefinition?: ContainerDefinition;
  definition?: ContainerDefinition;
  expandDefinition: ContainerDefinition;
  withNarrowVertical?: boolean;
  withTitle?: boolean;
}

const Container = ({
  allowChildPanelExpand = true,
  allowExpand = false,
  definition,
  layoutDefinition,
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
        allowPanelExpand={allowChildPanelExpand}
        children={
          definition
            ? definition.children
            : layoutDefinition
            ? layoutDefinition.children
            : []
        }
      />
    </Grid>
  );

  // return (
  //   <LayoutPanel
  //     allowExpand={allowExpand}
  //     className="relative"
  //     definition={panelDefinition}
  //     events={{
  //       onMouseEnter: allowExpand
  //         ? () => {
  //             setShowZoomIcon(true);
  //           }
  //         : undefined,
  //
  //       onMouseLeave: allowExpand
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
  //       allowPanelExpand={allowChildPanelExpand}
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
