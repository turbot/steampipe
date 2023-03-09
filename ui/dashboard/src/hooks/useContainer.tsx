import {
  createContext,
  ReactNode,
  useCallback,
  useContext,
  useEffect,
  useState,
} from "react";
import { noop } from "../utils/func";
import { PanelDefinition } from "../types";

export type ContainerChildVisibility = "visible" | "hidden";

type ContainerChildrenVisibility = {
  [key in ContainerChildVisibility]: PanelDefinition[];
};

type IContainerContext = {
  childVisibility?: ContainerChildrenVisibility;
  showTitle: boolean;
  updateChildStatus: (
    panel: PanelDefinition,
    visibility: ContainerChildVisibility
  ) => void;
};

type ContainerProviderProps = {
  children: ReactNode;
};

const ContainerContext = createContext<IContainerContext | null>({
  showTitle: true,
  updateChildStatus: noop,
});

const ContainerProvider = ({ children }: ContainerProviderProps) => {
  const [childVisibility, setChildVisibility] =
    useState<ContainerChildrenVisibility>({
      visible: [],
      hidden: [],
    });
  const [showTitle, setShowTitle] = useState(false);

  const updateChildStatus = useCallback(
    (child: PanelDefinition, visibility: ContainerChildVisibility) => {
      // Determine if the child is already marked as visible or hidden and update
      // state if required to force a re-render of the container's title

      const visibleIndex = childVisibility.visible.findIndex(
        (p) => p.name === child.name
      );
      const hiddenIndex = childVisibility.hidden.findIndex(
        (p) => p.name === child.name
      );
      if (visibility === "visible") {
        // If it's already marked as visible, nothing to do
        if (visibleIndex > -1 && hiddenIndex === -1) {
          return;
        }
        // Add to visible list if required
        const newVisible =
          visibleIndex > -1
            ? childVisibility.visible
            : [...childVisibility.visible, child];
        // Remove from hidden list if required
        const newHidden =
          hiddenIndex > -1
            ? [
                ...childVisibility.hidden.slice(0, hiddenIndex),
                ...childVisibility.hidden.slice(
                  hiddenIndex + 1,
                  childVisibility.hidden.length - 1
                ),
              ]
            : childVisibility.hidden;
        setChildVisibility({
          visible: newVisible,
          hidden: newHidden,
        });
      } else if (visibility === "hidden") {
        // If it's already marked as hidden, nothing to do
        if (hiddenIndex > -1 && visibleIndex === -1) {
          return;
        }
        // Remove from visible list if required
        const newVisible =
          visibleIndex > -1
            ? [
                ...childVisibility.visible.slice(0, visibleIndex),
                ...childVisibility.visible.slice(
                  visibleIndex + 1,
                  childVisibility.visible.length - 1
                ),
              ]
            : childVisibility.visible;
        // Add to hidden list if required
        const newHidden =
          hiddenIndex > -1
            ? childVisibility.hidden
            : [...childVisibility.hidden, child];
        setChildVisibility({
          visible: newVisible,
          hidden: newHidden,
        });
      }
    },
    [childVisibility, setChildVisibility]
  );

  useEffect(() => {
    if (
      showTitle &&
      childVisibility.hidden.length > 0 &&
      childVisibility.visible.length === 0
    ) {
      setShowTitle(false);
    } else if (!showTitle && childVisibility.visible.length > 0) {
      setShowTitle(true);
    }
  }, [childVisibility, showTitle, setShowTitle]);

  return (
    <ContainerContext.Provider
      value={{
        childVisibility,
        showTitle,
        updateChildStatus,
      }}
    >
      {children}
    </ContainerContext.Provider>
  );
};

const useContainer = () => {
  const context = useContext(ContainerContext);
  if (context === undefined) {
    // Normally we'd throw an error here, but we may not be in the context of
    // a container, so I'll just send some sensible defaults
    return {
      showTitle: true,
      updateChildStatus: noop,
    };
  }
  return context as IContainerContext;
};

export { ContainerContext, ContainerProvider, useContainer };
