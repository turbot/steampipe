import CheckPanel from "../CheckPanel";
import sortBy from "lodash/sortBy";
import useMediaMode from "../../../../hooks/useMediaMode";
import {
  CheckGroupingActions,
  useCheckGrouping,
} from "../../../../hooks/useCheckGrouping";
import { CheckNode } from "../common";
import { useEffect } from "react";

interface CheckGroupingProps {
  node: CheckNode;
}

const CheckGrouping = ({ node }: CheckGroupingProps) => {
  const { dispatch, nodeStates } = useCheckGrouping();
  // const [restoreNodeStates, setRestoreNodeStates] =
  //   useState<CheckGroupStates | null>(null);
  const mediaMode = useMediaMode();

  // @ts-ignore
  // const previousStates = usePrevious({
  //   mediaMode,
  //   nodeStates,
  //   restoreNodeStates,
  // });

  // useEffect(() => {
  //   if (
  //     // @ts-ignore
  //     (!previousStates || previousStates.mediaMode === "screen") &&
  //     mediaMode === "print"
  //   ) {
  //     // @ts-ignore
  //     setRestoreNodeStates(previousStates.nodeStates);
  //     dispatch({ type: CheckGroupingActions.EXPAND_ALL_NODES });
  //   } else if (
  //     previousStates &&
  //     // @ts-ignore
  //     previousStates.mediaMode === "print" &&
  //     mediaMode === "screen"
  //   ) {
  //     // @ts-ignore
  //     console.log(previousStates.restoreNodeStates);
  //     dispatch({
  //       type: CheckGroupingActions.UPDATE_NODES,
  //       // @ts-ignore
  //       nodes: previousStates.restoreNodeStates,
  //     });
  //   }
  // }, [mediaMode, previousStates]);

  useEffect(() => {
    if (mediaMode === "print") {
      dispatch({ type: CheckGroupingActions.EXPAND_ALL_NODES });
    }
  }, [mediaMode]);

  return (
    <div className="space-y-4 md:space-y-6 col-span-12">
      {sortBy(node.children, "sort")?.map((child) => (
        <CheckPanel key={child.name} depth={1} node={child} />
      ))}
    </div>
  );
};

export default CheckGrouping;
