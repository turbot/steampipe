import CheckPanel from "../CheckPanel";
import { CheckNode } from "../common";

interface CheckGroupingProps {
  node: CheckNode;
}

const CheckGrouping = ({ node }: CheckGroupingProps) => {
  console.log(node.children);
  return (
    <div className="space-y-6 col-span-12">
      {node.children &&
        node.children.map((child) => (
          <CheckPanel key={child.name} node={child} />
        ))}
    </div>
  );
};

export default CheckGrouping;
