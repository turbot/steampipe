import CheckGroupingEditor from "../CheckGroupingEditor";
import Icon from "../../../Icon";
import useCheckGroupingConfig from "../../../../hooks/useCheckGroupingConfig";
import { CheckDisplayGroup } from "../common";
import { classNames } from "../../../../utils/styles";
import { ReactNode, useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";

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
          <span className="text-foreground-lighter">=</span>
          <span className="font-medium">{item.value}</span>
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
  const [isValid, setIsValid] = useState(false);
  const [_, setSearchParams] = useSearchParams();
  const groupingConfig = useCheckGroupingConfig();
  const [modifiedConfig, setModifiedConfig] =
    useState<CheckDisplayGroup[]>(groupingConfig);

  useEffect(() => {
    const isValid = modifiedConfig.every((c) => {
      switch (c.type) {
        case "benchmark":
        case "control":
        case "result":
        case "reason":
        case "resource":
        case "severity":
        case "status":
          return !c.value;
        case "dimension":
        case "tag":
          return !!c.value;
      }
    });
    setIsValid(isValid);
  }, [modifiedConfig, setIsValid]);

  const saveGroupingConfig = (toSave) => {
    setSearchParams((previous) => ({
      ...previous,
      grouping: toSave
        .map((c) =>
          c.type === "dimension" || c.type === "tag"
            ? `${c.type}|${c.value}`
            : c.type,
        )
        .join(","),
    }));
  };

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
        {!showEditor && (
          <Icon
            className="h-5 w-5 cursor-pointer"
            icon="edit_square"
            onClick={() => setShowEditor(true)}
            title="Edit grouping"
          />
        )}
        {showEditor && (
          <>
            <Icon
              className="h-5 w-5 font-medium cursor-pointer"
              icon="close"
              onClick={() => setShowEditor(false)}
              title="Cancel changes"
            />
            <Icon
              className={classNames(
                "h-5 w-5 font-medium",
                isValid
                  ? "text-ok cursor-pointer"
                  : "text-foreground-lighter cursor-not-allowed",
              )}
              icon="done"
              onClick={() => {
                setShowEditor(false);
                saveGroupingConfig(modifiedConfig);
              }}
              title={isValid ? "Save changes" : "Invalid grouping config"}
            />
          </>
        )}
      </div>
      {showEditor && (
        <CheckGroupingEditor
          config={modifiedConfig}
          setConfig={setModifiedConfig}
        />
      )}
    </>
  );
};

export default CheckGroupingConfig;
