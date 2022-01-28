import Select from "react-select";
import { getColumnIndex } from "../../../../utils/data";
import { IInput, InputProps } from "../index";
import { ThemeNames, useTheme } from "../../../../hooks/useTheme";
import { useEffect, useMemo, useState } from "react";

type SelectInputProps = InputProps & {
  multi?: boolean;
};

const SelectInput = (props: SelectInputProps) => {
  const [_, setRandomVal] = useState(0);
  const { theme, wrapperRef } = useTheme();
  const options = useMemo(() => {
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
    control: (provided, state) => {
      return {
        ...provided,
        backgroundColor:
          theme.name === ThemeNames.STEAMPIPE_DARK ? blackScale2 : background,
        borderColor: state.isFocused ? "#2684FF" : blackScale3,
        boxShadow: "none",
      };
    },
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
        options={options}
        placeholder={
          (props.properties && props.properties.placeholder) ||
          "Please select..."
        }
        styles={customStyles}
      />
    </form>
  );
};

export default SelectInput;
