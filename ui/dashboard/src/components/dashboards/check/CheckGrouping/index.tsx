import CheckPanel from "../CheckPanel";
import sortBy from "lodash/sortBy";
import {
  CheckGroupNodeStates,
  CheckGroupingActions,
  useCheckGrouping,
} from "../../../../hooks/useCheckGrouping";
import { CheckNode } from "../common";
import { useCallback, useEffect, useState } from "react";

interface CheckGroupingProps {
  node: CheckNode;
}

const CheckGrouping = ({ node }: CheckGroupingProps) => {
  const { dispatch, nodeStates } = useCheckGrouping();
  const [restoreNodeStates, setRestoreNodeStates] =
    useState<CheckGroupNodeStates | null>(null);

  const expand = useCallback(() => {
    setRestoreNodeStates(nodeStates);
    dispatch({ type: CheckGroupingActions.EXPAND_ALL_NODES });
  }, [dispatch, nodeStates]);

  const restore = useCallback(() => {
    if (restoreNodeStates) {
      dispatch({
        type: CheckGroupingActions.UPDATE_NODES,
        nodes: restoreNodeStates,
      });
    }
  }, [dispatch, restoreNodeStates]);

  useEffect(() => {
    window.onbeforeprint = expand;
    window.onafterprint = restore;
  }, [expand, restore]);

  return (
    <div className="space-y-4 md:space-y-6 col-span-12">
      {sortBy(node.children, "sort")?.map((child) => (
        <CheckPanel key={child.name} depth={1} node={child} />
      ))}
    </div>
  );
};

export default CheckGrouping;
