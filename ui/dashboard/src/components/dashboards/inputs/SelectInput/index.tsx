import Select from "react-select";
import useSelectInputStyles from "./useSelectInputStyles";
import { getColumnIndex } from "../../../../utils/data";
import { InputProps } from "../index";
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

const SelectInput = (props: SelectInputProps) => {
  const { dispatch, selectedDashboardInputs } = useDashboard();
  const [initialisedFromState, setInitialisedFromState] = useState(false);
  const [value, setValue] = useState<SelectOption | SelectOption[] | null>(
    null
  );
  const options: SelectOption[] = useMemo(() => {
    if (!props.data || !props.data.columns || !props.data.rows) {
      return [];
    }
    const labelColIndex = getColumnIndex(props.data.columns, "label");
    const valueColIndex = getColumnIndex(props.data.columns, "value");

    if (labelColIndex === -1 || valueColIndex === -1) {
      return [];
    }

    return props.data.rows.map((row) => ({
      label: row[labelColIndex],
      value: row[valueColIndex],
    }));
  }, [props.data]);

  useEffect(() => {
    // If we've already set a value...
    if (initialisedFromState) {
      return;
    }

    const stateValue = selectedDashboardInputs[props.name];

    if (!stateValue) {
      setInitialisedFromState(true);
      return;
    }

    // If we haven't got the data we need yet...
    if (!options || options.length === 0) {
      return;
    }

    const parsedUrlValue = props.multi ? stateValue.split(",") : stateValue;

    const foundOption = props.multi
      ? options.filter((option) =>
          option.value
            ? parsedUrlValue.indexOf(option.value.toString()) >= 0
            : false
        )
      : options.find((option) =>
          option.value ? option.value.toString() === parsedUrlValue : false
        );

    if (!foundOption) {
      setInitialisedFromState(true);
      return;
    }

    setValue(foundOption);
    setInitialisedFromState(true);
  }, [
    props.name,
    props.multi,
    initialisedFromState,
    selectedDashboardInputs,
    options,
  ]);

  useEffect(() => {
    if (!initialisedFromState) {
      return;
    }

    const stateValue = selectedDashboardInputs[props.name];

    // @ts-ignore
    if ((!value || value.length === 0) && stateValue) {
      dispatch({ type: "delete_dashboard_input", name: props.name });
      return;
    }

    if (props.multi) {
      if (value) {
        // @ts-ignore
        const desiredValue = value.map((v) => v.value).join(",");
        if (stateValue !== desiredValue) {
          dispatch({
            type: "set_dashboard_input",
            name: props.name,
            value: desiredValue,
          });
        }
      }
    } else {
      // @ts-ignore
      if (value && stateValue !== value.value) {
        dispatch({
          type: "set_dashboard_input",
          name: props.name,
          // @ts-ignore
          value: value.value,
        });
      }
    }
  }, [
    dispatch,
    props.name,
    props.multi,
    initialisedFromState,
    selectedDashboardInputs,
    value,
  ]);

  const styles = useSelectInputStyles();

  if (!styles) {
    return null;
  }

  return (
    <form>
      {props.title && (
        <label
          className="block mb-1 text-sm"
          id={`${props.name}.label`}
          htmlFor={`${props.name}.input`}
        >
          {props.title}
        </label>
      )}
      <Select
        aria-labelledby={`${props.name}.input`}
        className="basic-single"
        classNamePrefix="select"
        menuPortalTarget={document.body}
        inputId={`${props.name}.input`}
        isDisabled={!props.data}
        isLoading={!props.data}
        isClearable
        isRtl={false}
        isSearchable
        isMulti={props.multi}
        // menuIsOpen
        name={props.name}
        // @ts-ignore
        onChange={(value) => setValue(value)}
        options={options}
        placeholder={
          (props.properties && props.properties.placeholder) ||
          "Please select..."
        }
        styles={styles}
        value={value}
      />
    </form>
  );
};

export default SelectInput;
