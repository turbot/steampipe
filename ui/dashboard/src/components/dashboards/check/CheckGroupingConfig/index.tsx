import CheckGroupingEditor from "../CheckGroupingEditor";
import Icon from "../../../Icon";
import useCheckGroupingConfig from "../../../../hooks/useCheckGroupingConfig";
import { CheckDisplayGroup } from "../common";
import { ReactNode, useState } from "react";
import { useDashboardControls } from "../../layout/Dashboard/DashboardControlsProvider";

type CheckGroupingTitleLabelProps = {
  item: CheckDisplayGroup;
};

const CheckGroupingTitleLabel = ({ item }: CheckGroupingTitleLabelProps) => {
  switch (item.type) {
    case "dimension":
    case "tag":
      return (
        <div className="space-x-1">
          <span className="capitalize font-medium">{item.type}</span>
          <span>=</span>
          <span>{item.value}</span>
        </div>
      );
    default:
      return (
        <div>
          <span className="capitalize font-medium">{item.type}</span>
        </div>
      );
  }
};

const CheckGroupingConfig = () => {
  const [showEditor, setShowEditor] = useState(false);
  const groupingConfig = useCheckGroupingConfig();
  return (
    <>
      <div className="flex items-center space-x-3 shrink-0">
        <Icon className="h-5 w-5" icon="workspaces" />
        {groupingConfig
          .map<ReactNode>((item) => (
            <CheckGroupingTitleLabel
              key={`${item.type}${!!item.value ? `-${item.value}` : ""}`}
              item={item}
            />
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
      {showEditor && <CheckGroupingEditor config={groupingConfig} />}
    </>
  );
};

export default CheckGroupingConfig;
