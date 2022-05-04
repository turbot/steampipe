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
  const mediaMode = useMediaMode();
  useEffect(() => {
    if (mediaMode === "print") {
      console.log("expanding");
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
