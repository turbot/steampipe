import CheckPanel from "../CheckPanel";
import { CheckNode } from "../common";

interface CheckGroupingProps {
  node: CheckNode;
}

const CheckGrouping = ({ node }: CheckGroupingProps) => {
  return (
    <div className="space-y-8 col-span-12">
      {node.children?.map((child) => (
        <CheckPanel key={child.name} node={child} />
      ))}
    </div>
  );
};

export default CheckGrouping;
