import CheckFilterEditor from "../CheckFilterEditor";
import Icon from "../../../Icon";
import useCheckFilterConfig from "../../../../hooks/useCheckFilterConfig";
import { CheckFilter } from "../common";
import { ReactNode, useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { classNames } from "../../../../utils/styles";

type CheckFilterTitleLabelProps = {
  item: CheckFilter;
};

const CheckFilterTitleLabel = ({ item }: CheckFilterTitleLabelProps) => {
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

const CheckFilterConfig = () => {
  const [showEditor, setShowEditor] = useState(false);
  const [isValid, setIsValid] = useState(false);
  const [_, setSearchParams] = useSearchParams();
  const filterConfig = useCheckFilterConfig();
  const [modifiedConfig, setModifiedConfig] =
    useState<CheckFilter[]>(filterConfig);

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

  const saveFilterConfig = (toSave) => {
    setSearchParams((previous) => ({
      ...previous,
      filter: toSave
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
        <Icon className="h-5 w-5" icon="filter_list" />
        {filterConfig
          .map<ReactNode>((item) => (
            <CheckFilterTitleLabel
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
            title="Edit filter"
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
                saveFilterConfig(modifiedConfig);
              }}
              title={isValid ? "Save changes" : "Invalid filter config"}
            />
          </>
        )}
      </div>
      {showEditor && (
        <CheckFilterEditor
          config={modifiedConfig}
          setConfig={setModifiedConfig}
        />
      )}
    </>
  );
};

export default CheckFilterConfig;
