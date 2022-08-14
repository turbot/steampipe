import CreatableSelect from "react-select/creatable";
import useSelectInputStyles from "../common/useSelectInputStyles";
import { DashboardActions, useDashboard } from "../../../../hooks/useDashboard";
import { getColumn } from "../../../../utils/data";
import { InputProps } from "../types";
import {
  MultiValueLabelWithTags,
  OptionWithTags,
  SingleValueWithTags,
} from "../common/Common";
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

const findOptions = (options, multi, value) => {
  return multi
    ? options.filter((option) =>
        option.value ? value.indexOf(option.value.toString()) >= 0 : false
      )
    : options.find((option) =>
        option.value ? option.value.toString() === value : false
      );
};

const SelectInput = ({
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
    // console.log({
    //   name,
    //   status,
    //   multi,
    //   options,
    //   initialisedFromState,
    //   placeholder: properties.placeholder,
    //   stateValue,
    //   value,
    // });
    // console.log(name, status, options);
    // If we haven't got the data we need yet...
    if (
      // This property is only present in workspaces >=v0.16.x
      (status !== undefined && status !== "complete") ||
      !options ||
      options.length === 0
    ) {
      return;
    }
    //
    // if (initialisedFromState) {
    //   return;
    // }

    // If this is first load, and we have a value from state, initialise it
    if (!initialisedFromState && stateValue) {
      const parsedUrlValue = multi ? stateValue.split(",") : stateValue;
      const foundOptions = findOptions(options, multi, parsedUrlValue);
      setValue(foundOptions || null);
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
      const foundOptions = findOptions(options, multi, parsedUrlValue);
      setValue(foundOptions || null);
      // dispatch({
      //   type: DashboardActions.SET_DASHBOARD_INPUT,
      //   name,
      //   value: getValueForState(multi, foundOptions),
      //   recordInputsHistory: false,
      // });
    } else if (initialisedFromState && !stateValue) {
      if (properties.placeholder) {
        // console.log("Value cleared in state and no placeholder");
        setValue(null);
      } else {
        // console.log(
        //   "Value cleared in state and placeholder so choosing first item"
        // );
        const newValue = multi ? [options[0]] : options[0];
        // console.log(name, status, newValue);
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

  // useEffect(() => {
  //   if (!initialisedFromState) {
  //     return;
  //   }
  //
  //   // If we don't have a value in state and in the control, nothing to do
  //   if (!stateValue && !value) {
  //     // console.log("value changed - nothing to do");
  //     return;
  //   }
  //
  //   // const parsedStateValue = getValueFromState(multi, stateValue);
  //   // const parsedOptionValue = multi
  //   //   ? value
  //   //     ? // @ts-ignore
  //   //       value.map((v) => v.value)
  //   //     : []
  //   //   : // @ts-ignore
  //   //     value.value;
  //   // const sameValue = multi
  //   //   ? JSON.stringify(parsedStateValue) === JSON.stringify(parsedOptionValue)
  //   //   : parsedStateValue === parsedOptionValue;
  //   //
  //   // // If nothing has changed, don't dispatch changes
  //   // if (sameValue) {
  //   //   console.log("value changed - no change as same value", {
  //   //     name,
  //   //     multi,
  //   //     stateValue,
  //   //     value,
  //   //     parsedStateValue,
  //   //     parsedOptionValue,
  //   //   });
  //   //   return;
  //   // }
  //
  //   // @ts-ignore
  //   if (!value || value.length === 0) {
  //     // console.log("value changed - deleting", {
  //     //   initialisedFromState,
  //     //   multi,
  //     //   name,
  //     //   stateValue,
  //     //   // parsedStateValue,
  //     //   // parsedOptionValue,
  //     //   value,
  //     // });
  //     dispatch({
  //       type: DashboardActions.DELETE_DASHBOARD_INPUT,
  //       name,
  //       recordInputsHistory: true,
  //     });
  //     if (properties.placeholder) {
  //       // console.log("Value cleared in state and no placeholder");
  //       setValue(null);
  //     } else {
  //       // console.log(
  //       //   "Value cleared in state and placeholder so choosing first item"
  //       // );
  //       setValue(multi ? [options[0]] : options[0]);
  //     }
  //     return;
  //   }
  //
  //   // console.log("value changed - setting", {
  //   //   initialisedFromState,
  //   //   multi,
  //   //   name,
  //   //   stateValue,
  //   //   // parsedStateValue,
  //   //   // parsedOptionValue,
  //   //   value,
  //   // });
  //   dispatch({
  //     type: DashboardActions.SET_DASHBOARD_INPUT,
  //     name,
  //     value: getValueForState(multi, value),
  //     recordInputsHistory: true,
  //   });
  // }, [
  //   dispatch,
  //   initialisedFromState,
  //   multi,
  //   name,
  //   properties.placeholder,
  //   stateValue,
  //   value,
  // ]);

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
        menuPortalTarget={document.body}
        inputId={`${name}.input`}
        isDisabled={(!properties.options && !data) || dataMode === "snapshot"}
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

export default SelectInput;
