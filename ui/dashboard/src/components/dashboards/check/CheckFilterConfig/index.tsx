import CheckFilterEditor from "../CheckFilterEditor";
import get from "lodash/get";
import Icon from "../../../Icon";
import useCheckFilterConfig from "../../../../hooks/useCheckFilterConfig";
import { AndFilter, CheckFilter, Filter, OrFilter } from "../common";
import { classNames } from "../../../../utils/styles";
import { Fragment, ReactNode, useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";

const filtersToText = (filter) => {
  if ("type" in filter) {
    // Convert filter to text
    let textParts: ReactNode[] = [];
    if (filter.key) {
      textParts.push(<span>{filter.key}</span>);
    } else {
      textParts.push(<span className="capitalize">{filter.type}</span>);
    }
    textParts.push(<span>{filter.value}</span>);

    return (
      <span className="space-x-1">
        {textParts.map((item, index) => (
          <Fragment key={index}>
            {!!index && <span className="text-foreground-lighter">=</span>}
            {item}
          </Fragment>
        ))}
      </span>
    );
  } else if ("or" in filter) {
    // Or filter group
    return filter.or.map((item, index) => (
      <Fragment key={index}>
        {!!index && <span className="text-foreground-lighter">or</span>}
        {filtersToText(item)}
      </Fragment>
    ));
  } else if ("and" in filter) {
    // And filter group
    return filter.and.map((item, index) => (
      <Fragment key={index}>
        {!!index && <span className="text-foreground-lighter">and</span>}
        {filtersToText(item)}
      </Fragment>
    ));
    // return `(${filter.and.map(filtersToText).join(" AND ")})`;
  }
};

const validateFilter = (filter: Filter): boolean => {
  return (
    !!filter.type && (filter.key !== undefined || filter.value !== undefined)
  );
};

const validateOrFilter = (orFilter: OrFilter): boolean => {
  return Array.isArray(orFilter.or) && orFilter.or.every(validateFilter);
};

const validateAndFilter = (andFilter: AndFilter): boolean => {
  return (
    Array.isArray(andFilter.and) &&
    andFilter.and.every((item) => {
      if ("or" in item) {
        return validateOrFilter(item);
      } else {
        return validateFilter(item as Filter);
      }
    })
  );
};

const validateCheckFilter = (checkFilter: CheckFilter): boolean => {
  if (checkFilter.and) {
    return validateAndFilter(checkFilter);
  } else if (checkFilter.or) {
    return validateOrFilter(checkFilter);
  } else {
    throw new Error("Invalid check filter.");
  }
};

const CheckFilterConfig = () => {
  const [showEditor, setShowEditor] = useState(false);
  const [isValid, setIsValid] = useState(false);
  const [_, setSearchParams] = useSearchParams();
  const filterConfig = useCheckFilterConfig();
  const [modifiedConfig, setModifiedConfig] =
    useState<CheckFilter>(filterConfig);

  useEffect(() => {
    if (!modifiedConfig) {
      setIsValid(true);
      return;
    }

    if (!!modifiedConfig.and) {
      setIsValid(validateAndFilter(modifiedConfig));
      return;
    } else if (!!modifiedConfig.or) {
      setIsValid(validateOrFilter(modifiedConfig));
      return;
    }
    setIsValid(false);
  }, [modifiedConfig, setIsValid]);

  const saveFilterConfig = (toSave: CheckFilter) => {
    setSearchParams((previous) => {
      if (toSave.and && toSave.and.length === 0) {
        const newConfig = {
          ...previous,
        };
        delete newConfig["where"];
        return newConfig;
      } else {
        const asJson = JSON.stringify(toSave);
        return {
          ...previous,
          where: asJson,
        };
      }
    });
  };

  return (
    <>
      <div className="flex items-center space-x-3 shrink-0">
        <Icon className="h-5 w-5" icon="filter_list" />
        {get(filterConfig, "and", []).length > 0 && (
          <div className="space-x-2">{filtersToText(filterConfig)}</div>
        )}
        {get(filterConfig, "and", []).length === 0 &&
          get(filterConfig, "or", []).length === 0 && (
            <span className="text-foreground-lighter">No filters</span>
          )}
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

export {
  validateCheckFilter,
  validateFilter,
  validateAndFilter,
  validateOrFilter,
};
