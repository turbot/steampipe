import Select from "react-select";
import { getColumnIndex } from "../../../../utils/data";
import { InputProps } from "../index";
import { ThemeNames, useTheme } from "../../../../hooks/useTheme";
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
  const [, setRandomVal] = useState(0);
  const { theme, wrapperRef } = useTheme();
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
      ? options.filter((option) => parsedUrlValue.indexOf(option.value) >= 0)
      : options.find((option) => option.value === parsedUrlValue);

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

  // This is annoying, but unless I force a refresh the theme doesn't stay in sync when you switch
  useEffect(() => setRandomVal(Math.random()), [theme.name]);

  if (!wrapperRef) {
    return null;
  }

  // @ts-ignore
  const style = window.getComputedStyle(wrapperRef);
  const background = style.getPropertyValue("--color-background");
  const foreground = style.getPropertyValue("--color-foreground");
  const blackScale1 = style.getPropertyValue("--color-black-scale-1");
  const blackScale2 = style.getPropertyValue("--color-black-scale-2");
  const blackScale3 = style.getPropertyValue("--color-black-scale-3");

  const customStyles = {
    clearIndicator: (provided) => ({
      ...provided,
      cursor: "pointer",
    }),
    control: (provided, state) => {
      return {
        ...provided,
        backgroundColor:
          theme.name === ThemeNames.STEAMPIPE_DARK ? blackScale2 : background,
        borderColor: state.isFocused ? "#2684FF" : blackScale3,
        boxShadow: "none",
      };
    },
    dropdownIndicator: (provided) => ({
      ...provided,
      cursor: "pointer",
    }),
    singleValue: (provided) => {
      return {
        ...provided,
        color: foreground,
      };
    },
    menu: (provided) => {
      return {
        ...provided,
        backgroundColor:
          theme.name === ThemeNames.STEAMPIPE_DARK ? blackScale2 : background,
        border: `1px solid ${blackScale3}`,
        boxShadow: "none",
        marginTop: 0,
        marginBottom: 0,
      };
    },
    menuList: (provided) => {
      return {
        ...provided,
        paddingTop: 0,
        paddingBottom: 0,
      };
    },
    option: (provided, state) => {
      return {
        ...provided,
        backgroundColor: state.isFocused ? blackScale1 : "none",
        color: foreground,
      };
    },
  };

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
        styles={customStyles}
        value={value}
      />
    </form>
  );
};

export default SelectInput;
