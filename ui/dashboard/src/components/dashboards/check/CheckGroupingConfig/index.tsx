import Icon from "../../../Icon";
import useCheckGroupingConfig from "../../../../hooks/useCheckGroupingConfig";
import { CheckDisplayGroup } from "../common";
import { CheckGroupingEditorModal } from "../CheckGroupingEditorModal";
import { useState } from "react";

type CheckGroupingTitleLabelProps = {
  item: CheckDisplayGroup;
};

const CheckGroupingTitleLabel = ({ item }: CheckGroupingTitleLabelProps) => {
  switch (item.type) {
    case "dimension":
    case "tag":
      return (
        <div className="space-x-1">
          <span className="capitalize">{item.type}</span>
          <span>=</span>
          <span>{item.value}</span>
        </div>
      );
    default:
      return (
        <div>
          <span className="capitalize">{item.type}</span>
        </div>
      );
  }
};

const CheckGroupingConfig = ({}) => {
  const [showEditor, setShowEditor] = useState(false);
  const groupingConfig = useCheckGroupingConfig();
  return (
    <>
      <div className="flex items-center space-x-2 shrink-0">
        <span className="font-medium">Grouping:</span>
        {groupingConfig
          .map<React.ReactNode>((item) => (
            <CheckGroupingTitleLabel key={item.id} item={item} />
          ))
          .reduce((prev, curr, idx) => [
            prev,
            <Icon key={idx} className="h-4 w-4" icon="arrow-long-right" />,
            curr,
          ])}
        <Icon
          className="h-4 w-4 text-foreground-light hover:text-foreground"
          icon="pencil-square"
          onClick={() => setShowEditor(true)}
          title="Edit grouping"
        />
      </div>
      {showEditor && (
        <CheckGroupingEditorModal
          config={groupingConfig}
          setShowEditor={setShowEditor}
        />
      )}
    </>
  );
};

export default CheckGroupingConfig;
