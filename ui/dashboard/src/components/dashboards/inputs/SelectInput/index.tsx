import Select, {
  components,
  OptionProps,
  SingleValueProps,
} from "react-select";
import useSelectInputStyles from "./useSelectInputStyles";
import { ColorGenerator } from "../../../../utils/color";
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

const stringColorMap = {};
const colorGenerator = new ColorGenerator(24, 4);

const stringToColour = (str) => {
  if (stringColorMap[str]) {
    return stringColorMap[str];
  }
  const color = colorGenerator.nextColor().hex;
  stringColorMap[str] = color;
  return color;
};

const OptionTag = ({ tagKey, tagValue }) => (
  <span
    className="rounded-md text-xs"
    style={{ color: stringToColour(tagValue) }}
    title={`${tagKey} = ${tagValue}`}
  >
    {tagValue}
  </span>
);

const LabelTagWrapper = ({ label, tags }) => (
  <div className="space-x-2">
    {/*@ts-ignore*/}
    <span>{label}</span>
    {/*@ts-ignore*/}
    {Object.entries(tags || {}).map(([tagKey, tagValue]) => (
      <OptionTag key={tagKey} tagKey={tagKey} tagValue={tagValue} />
    ))}
  </div>
);

const OptionWithTags = (props: OptionProps) => (
  <components.Option {...props}>
    {/*@ts-ignore*/}
    <LabelTagWrapper label={props.data.label} tags={props.data.tags} />
  </components.Option>
);

const SingleValueWithTags = ({ children, ...props }: SingleValueProps) => {
  return (
    <components.SingleValue {...props}>
      {/*@ts-ignore*/}
      <LabelTagWrapper label={props.data.label} tags={props.data.tags} />
    </components.SingleValue>
  );
};

const MultiValueLabelWithTags = ({ children, ...props }: SingleValueProps) => {
  return (
    <components.MultiValueLabel {...props}>
      {/*@ts-ignore*/}
      <LabelTagWrapper label={props.data.label} tags={props.data.tags} />
    </components.MultiValueLabel>
  );
};

const SelectInput = ({ data, multi, name, properties }: SelectInputProps) => {
  const { dispatch, selectedDashboardInputs } = useDashboard();
  const [initialisedFromState, setInitialisedFromState] = useState(false);
  const [value, setValue] = useState<SelectOption | SelectOption[] | null>(
    null
  );

  // Get the options for the select
  const options: SelectOption[] = useMemo(() => {
    // If no options defined at all
    if (
      (!properties?.options || properties?.options.length === 0) &&
      (!data || !data.columns || !data.rows)
    ) {
      return [];
    }

    if (data) {
      const labelColIndex = getColumnIndex(data.columns, "label");
      const valueColIndex = getColumnIndex(data.columns, "value");
      const tagsColIndex = getColumnIndex(data.columns, "tags");

      if (labelColIndex === -1 || valueColIndex === -1) {
        return [];
      }

      return data.rows.map((row) => ({
        label: row[labelColIndex],
        value: row[valueColIndex],
        tags: tagsColIndex > -1 ? row[tagsColIndex] : {},
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
  }, [properties.options, data]);

  const stateValue = selectedDashboardInputs[name];

  // Bind the selected option to the reducer state
  useEffect(() => {
    // If we haven't got the data we need yet...
    if (!options || options.length === 0) {
      return;
    }

    // If this is first load and we have a value from state, initialise it
    if (!initialisedFromState && stateValue) {
      const parsedUrlValue = multi ? stateValue.split(",") : stateValue;

      const foundOption = multi
        ? options.filter((option) =>
            option.value
              ? parsedUrlValue.indexOf(option.value.toString()) >= 0
              : false
          )
        : options.find((option) =>
            option.value ? option.value.toString() === parsedUrlValue : false
          );
      setValue(foundOption || null);
      setInitialisedFromState(true);
    } else if (!initialisedFromState && !stateValue) {
      setInitialisedFromState(true);
    } else if (initialisedFromState && !stateValue) {
      setValue(null);
    }
  }, [initialisedFromState, name, options, stateValue]);

  useEffect(() => {
    if (!initialisedFromState) {
      return;
    }

    // @ts-ignore
    if ((!value || value.length === 0) && stateValue) {
      dispatch({ type: "delete_dashboard_input", name });
      return;
    }

    if (multi) {
      if (value) {
        // @ts-ignore
        const desiredValue = value.map((v) => v.value).join(",");
        if (stateValue !== desiredValue) {
          dispatch({
            type: "set_dashboard_input",
            name,
            value: desiredValue,
          });
        }
      }
    } else {
      // @ts-ignore
      if (value && stateValue !== value.value) {
        dispatch({
          type: "set_dashboard_input",
          name,
          // @ts-ignore
          value: value.value,
        });
      }
    }
  }, [dispatch, initialisedFromState, stateValue, value]);

  // useEffect(() => {
  //   // If we've already set a value...
  //   if (initialisedFromState) {
  //     return;
  //   }
  //
  //   const stateValue = selectedDashboardInputs[name];
  //
  //   if (!stateValue) {
  //     setInitialisedFromState(true);
  //     return;
  //   }
  //
  //   // If we haven't got the data we need yet...
  //   if (!options || options.length === 0) {
  //     return;
  //   }
  //
  //   const parsedUrlValue = multi ? stateValue.split(",") : stateValue;
  //
  //   const foundOption = multi
  //     ? options.filter((option) =>
  //         option.value
  //           ? parsedUrlValue.indexOf(option.value.toString()) >= 0
  //           : false
  //       )
  //     : options.find((option) =>
  //         option.value ? option.value.toString() === parsedUrlValue : false
  //       );
  //
  //   if (!foundOption) {
  //     setInitialisedFromState(true);
  //     return;
  //   }
  //
  //   setValue(foundOption);
  //   setInitialisedFromState(true);
  // }, [name, multi, initialisedFromState, selectedDashboardInputs, options]);
  //
  // useEffect(() => {
  //   if (!initialisedFromState) {
  //     return;
  //   }
  //
  //   const stateValue = selectedDashboardInputs[name];
  //
  //   // @ts-ignore
  //   if ((!value || value.length === 0) && stateValue) {
  //     dispatch({ type: "delete_dashboard_input", name });
  //     return;
  //   }
  //
  //   if (multi) {
  //     if (value) {
  //       // @ts-ignore
  //       const desiredValue = value.map((v) => v.value).join(",");
  //       if (stateValue !== desiredValue) {
  //         dispatch({
  //           type: "set_dashboard_input",
  //           name,
  //           value: desiredValue,
  //         });
  //       }
  //     }
  //   } else {
  //     // @ts-ignore
  //     if (value && stateValue !== value.value) {
  //       dispatch({
  //         type: "set_dashboard_input",
  //         name,
  //         // @ts-ignore
  //         value: value.value,
  //       });
  //     }
  //   }
  // }, [
  //   dispatch,
  //   name,
  //   multi,
  //   initialisedFromState,
  //   selectedDashboardInputs,
  //   value,
  // ]);

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
      <Select
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
        isDisabled={!properties.options && !data}
        isLoading={!properties.options && !data}
        isClearable
        isRtl={false}
        isSearchable
        isMulti={multi}
        // menuIsOpen
        name={name}
        // @ts-ignore
        onChange={(value) => setValue(value)}
        options={options}
        placeholder={
          (properties && properties.placeholder) || "Please select..."
        }
        styles={styles}
        value={value}
      />
    </form>
  );
};

export default SelectInput;
