import CreatableSelect from "react-select/creatable";
import useSelectInputStyles from "../common/useSelectInputStyles";
import { DashboardActions, DashboardDataModeLive } from "../../../../types";
import { getColumn } from "../../../../utils/data";
import { InputProps } from "../types";
import {
  MultiValueLabelWithTags,
  OptionWithTags,
  SingleValueWithTags,
} from "../common/Common";
import { useDashboard } from "../../../../hooks/useDashboard";
import { useEffect, useMemo, useState } from "react";

export interface SelectOption {
  label: string;
  value: string;
}

type SelectInputProps = InputProps & {
  multi?: boolean;
  name: string;
};

const getValueForState = (multi, option) => {
  if (multi) {
    // @ts-ignore
    return option.map((v) => v.value).join(",");
  } else {
    return option.value;
  }
};

const findOptionsForUrlValue = (
  options,
  multi,
  urlValue
): SelectOption | SelectOption[] => {
  // If we can't find any of the options in the data, we accept it, as this is a
  // combo box and the user can enter anything they like.
  if (multi) {
    const matchingOptions: SelectOption[] = [];
    for (const urlValuePart of urlValue) {
      const existingOption = options.find(
        (option) => option.value === urlValuePart
      );
      if (existingOption) {
        matchingOptions.push(existingOption);
      } else {
        matchingOptions.push({
          label: urlValuePart,
          value: urlValuePart,
        } as SelectOption);
      }
    }
    return matchingOptions;
  } else {
    const existingOption = options.find((option) => option.value === urlValue);
    if (existingOption) {
      return existingOption;
    } else {
      return {
        label: urlValue,
        value: urlValue,
      } as SelectOption;
    }
  }
};

const ComboInput = ({
  data,
  multi,
  name,
  properties,
  status,
}: SelectInputProps) => {
  const { dataMode, dispatch, selectedDashboardInputs } = useDashboard();
  const [initialisedFromState, setInitialisedFromState] = useState(false);
  const [value, setValue] = useState<SelectOption | SelectOption[] | null>(
    null
  );

  // Get the options for the select
  const options: SelectOption[] = useMemo(() => {
    // If no options defined at all
    if (
      ((!properties?.options || properties?.options.length === 0) &&
        (!data || !data.columns || !data.rows)) ||
      // This property is only present in workspaces >=v0.16.x
      (status !== undefined && status !== "complete")
    ) {
      return [];
    }

    if (data) {
      const labelCol = getColumn(data.columns, "label");
      const valueCol = getColumn(data.columns, "value");
      const tagsCol = getColumn(data.columns, "tags");

      if (!labelCol || !valueCol) {
        return [];
      }

      return data.rows.map((row) => ({
        label: row[labelCol.name],
        value: row[valueCol.name],
        tags: tagsCol ? row[tagsCol.name] : {},
      }));
    } else if (properties.options) {
      return properties.options.map((option) => ({
        label: option.label || option.name,
        value: option.name,
        tags: {},
      }));
    } else {
      return [];
    }
  }, [properties.options, data, status]);

  const stateValue = selectedDashboardInputs[name];

  // Bind the selected option to the reducer state
  useEffect(() => {
    // If we haven't got the data we need yet...
    if (
      // This property is only present in workspaces >=v0.16.x
      (status !== undefined && status !== "complete") ||
      !options ||
      options.length === 0
    ) {
      return;
    }

    // If this is first load, and we have a value from state, initialise it
    if (!initialisedFromState && stateValue) {
      const parsedUrlValue = multi ? stateValue.split(",") : stateValue;
      const foundOptions = findOptionsForUrlValue(
        options,
        multi,
        parsedUrlValue
      );
      setValue(foundOptions);
      setInitialisedFromState(true);
    } else if (!initialisedFromState && !stateValue && properties.placeholder) {
      setInitialisedFromState(true);
    } else if (
      !initialisedFromState &&
      !stateValue &&
      !properties.placeholder
    ) {
      setInitialisedFromState(true);
      const newValue = multi ? [options[0]] : options[0];
      setValue(newValue);
      dispatch({
        type: DashboardActions.SET_DASHBOARD_INPUT,
        name,
        value: getValueForState(multi, newValue),
        recordInputsHistory: false,
      });
    } else if (initialisedFromState && stateValue) {
      const parsedUrlValue = multi ? stateValue.split(",") : stateValue;
      const foundOptions = findOptionsForUrlValue(
        options,
        multi,
        parsedUrlValue
      );
      setValue(foundOptions);
    } else if (initialisedFromState && !stateValue) {
      if (properties.placeholder) {
        setValue(null);
      } else {
        const newValue = multi ? [options[0]] : options[0];
        setValue(newValue);
        dispatch({
          type: DashboardActions.SET_DASHBOARD_INPUT,
          name,
          value: getValueForState(multi, newValue),
          recordInputsHistory: false,
        });
      }
    }
  }, [
    dispatch,
    initialisedFromState,
    multi,
    name,
    options,
    properties.placeholder,
    stateValue,
    status,
  ]);

  const updateValue = (newValue) => {
    setValue(newValue);
    if (!newValue || newValue.length === 0) {
      dispatch({
        type: DashboardActions.DELETE_DASHBOARD_INPUT,
        name,
        recordInputsHistory: true,
      });
    } else {
      dispatch({
        type: DashboardActions.SET_DASHBOARD_INPUT,
        name,
        value: getValueForState(multi, newValue),
        recordInputsHistory: true,
      });
    }
  };

  const styles = useSelectInputStyles();

  if (!styles) {
    return null;
  }

  return (
    <form>
      {properties && properties.label && (
        <label
          className="block mb-1 text-sm"
          id={`${name}.label`}
          htmlFor={`${name}.input`}
        >
          {properties.label}
        </label>
      )}
      <CreatableSelect
        aria-labelledby={`${name}.input`}
        className="basic-single"
        classNamePrefix="select"
        components={{
          // @ts-ignore
          MultiValueLabel: MultiValueLabelWithTags,
          // @ts-ignore
          Option: OptionWithTags,
          // @ts-ignore
          SingleValue: SingleValueWithTags,
        }}
        createOptionPosition="first"
        formatCreateLabel={(inputValue) => `Use "${inputValue}"`}
        // @ts-ignore as this element definitely exists
        menuPortalTarget={document.getElementById("portals")}
        inputId={`${name}.input`}
        isDisabled={
          (!properties.options && !data) || dataMode !== DashboardDataModeLive
        }
        isLoading={!properties.options && !data}
        isClearable={!!properties.placeholder}
        isRtl={false}
        isSearchable
        isMulti={multi}
        // menuIsOpen
        name={name}
        // @ts-ignore
        onChange={updateValue}
        options={options}
        placeholder={
          properties && properties.placeholder ? properties.placeholder : null
        }
        styles={styles}
        value={value}
      />
    </form>
  );
};

export default ComboInput;
