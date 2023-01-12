import Child from "../Child";
import {
  ContainerDefinition,
  DashboardPanelType,
  PanelDefinition,
} from "../../../../types";
import { useDashboard } from "../../../../hooks/useDashboard";

type ChildrenProps = {
  children: ContainerDefinition[] | PanelDefinition[] | undefined;
  parentType: DashboardPanelType;
  showPanelControls?: boolean;
};

const Children = ({
  children = [],
  parentType,
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
            key={definition.name}
            layoutDefinition={child}
            panelDefinition={definition}
            parentType={parentType}
            showPanelControls={showPanelControls}
          />
        );
      })}
    </>
  );
};

export default Children;
