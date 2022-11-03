import Child from "../Child";
import { ContainerDefinition, PanelDefinition } from "../../../../types";
import { useDashboard } from "../../../../hooks/useDashboard";

type ChildrenProps = {
  children: ContainerDefinition[] | PanelDefinition[] | undefined;
  showPanelControls?: boolean;
};

const Children = ({
  children = [],
  showPanelControls = true,
}: ChildrenProps) => {
  const { panelsMap } = useDashboard();
  return (
    <>
      {children.map((child) => {
        const definition = panelsMap[child.name];
        if (!definition) {
          return null;
        }
        return (
          <Child
            layoutDefinition={child}
            panelDefinition={definition}
            panelsMap={panelsMap}
            showPanelControls={showPanelControls}
          />
        );
      })}
    </>
  );
};

export default Children;
