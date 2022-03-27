import Select, {
  components,
  MultiValueProps,
  OptionProps,
  SingleValueProps,
} from "react-select";
import useSelectInputStyles from "./useSelectInputStyles";
import { ColorGenerator } from "../../../../utils/color";
import { DashboardActions, useDashboard } from "../../../../hooks/useDashboard";
import { getColumnIndex } from "../../../../utils/data";
import { InputProperties, InputProps } from "../index";
import { useEffect, useMemo, useState } from "react";
import { WarningIcon } from "../../../../constants/icons";

interface SelectOptionTags {
  [key: string]: string;
}

export interface SelectOption {
  label: string;
  value: string;
  tags?: SelectOptionTags;
  artificial?: boolean;
}

type SelectInputProperties = InputProperties & {
  multi?: boolean;
};

type SelectInputProps = InputProps & {
  properties: SelectInputProperties;
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

const LabelTagWrapper = ({ artificial, label, tags }) => (
  <div className="space-x-2">
    {/*@ts-ignore*/}
    <span>{label}</span>
    {artificial && (
      <span
        className="text-yellow"
        title="Value is not in the defined options list"
      >
        <WarningIcon className="inline w-4 h-4 fill-yellow" />
      </span>
    )}
    {/*@ts-ignore*/}
    {Object.entries(tags || {}).map(([tagKey, tagValue]) => (
      <OptionTag key={tagKey} tagKey={tagKey} tagValue={tagValue} />
    ))}
  </div>
);

const OptionWithTags = (props: OptionProps) => (
  <components.Option {...props}>
    <LabelTagWrapper
      /*@ts-ignore*/
      artificial={props.data.artificial}
      /*@ts-ignore*/
      label={props.data.label}
      /*@ts-ignore*/
      tags={props.data.tags}
    />
  </components.Option>
);

const SingleValueWithTags = ({ children, ...props }: SingleValueProps) => {
  return (
    <components.SingleValue {...props}>
      <LabelTagWrapper
        /*@ts-ignore*/
        artificial={props.data.artificial}
        /*@ts-ignore*/
        label={props.data.label}
        /*@ts-ignore*/
        tags={props.data.tags}
      />
    </components.SingleValue>
  );
};

const MultiValueLabelWithTags = ({ children, ...props }: MultiValueProps) => {
  return (
    <components.MultiValueLabel {...props}>
      <LabelTagWrapper
        /*@ts-ignore*/
        artificial={props.data.artificial}
        /*@ts-ignore*/
        label={props.data.label}
        /*@ts-ignore*/
        tags={props.data.tags}
      />
    </components.MultiValueLabel>
  );
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

// const getInitialValue = (
//   stateValue: string | undefined,
//   inputDefault: string | undefined,
//   multi: boolean | undefined
// ): SelectOption | SelectOption[] | null => {
//   if (!inputDefault) {
//     return null;
//   }
//   if (multi) {
//     return inputDefault.split(",").map((value) => ({
//       label: value,
//       value,
//     }));
//   }
//   return {
//     label: inputDefault,
//     value: inputDefault,
//   };
// };

const addOptionIfNotExists = (
  options: SelectOption[],
  inputDefault: string | undefined,
  multi: boolean | undefined
): SelectOption[] => {
  if (!inputDefault) {
    return options;
  }

  const newOptions = [...options];

  if (multi) {
    const parts = inputDefault.split(",");
    for (const part of parts) {
      const matching = findOptions(options, false, part);
      if (!matching) {
        newOptions.push({
          label: part,
          value: part,
          artificial: true,
        });
      }
    }
  } else {
    const matching = findOptions(options, false, inputDefault);
    if (!matching) {
      newOptions.push({
        label: inputDefault,
        value: inputDefault,
        artificial: true,
      });
    }
  }
  return newOptions;
};

const SelectInput = ({ data, name, properties }: SelectInputProps) => {
  const { dispatch, selectedDashboardInputs } = useDashboard();
  const stateValue = selectedDashboardInputs[name];
  const [initialisedFromState, setInitialisedFromState] = useState(false);
  // const [value, setValue] = useState<SelectOption | SelectOption[] | null>(() =>
  //   getInitialValue(stateValue, properties.default, properties.multi)
  // );
  const [value, setValue] = useState<SelectOption | SelectOption[] | null>(
    null
  );

  // Get the options for the select
  const options: SelectOption[] = useMemo(() => {
    // If no options defined at all
    if (
      (!properties?.options || properties?.options.length === 0) &&
      !properties.default &&
      (!data || !data.columns || !data.rows)
    ) {
      return [];
    }

    let newOptions: SelectOption[] = [];

    if (data) {
      const labelColIndex = getColumnIndex(data.columns, "label");
      const valueColIndex = getColumnIndex(data.columns, "value");
      const tagsColIndex = getColumnIndex(data.columns, "tags");

      if (labelColIndex === -1 || valueColIndex === -1) {
        return [];
      }
      newOptions = data.rows.map((row) => ({
        label: row[labelColIndex],
        value: row[valueColIndex],
        tags: tagsColIndex > -1 ? row[tagsColIndex] : {},
      }));
    } else if (properties.options) {
      newOptions = properties.options.map((option) => ({
        label: option.label || option.name,
        value: option.name,
        tags: {},
      }));
    }
    newOptions = addOptionIfNotExists(newOptions, stateValue, properties.multi);
    return addOptionIfNotExists(
      newOptions,
      properties.default,
      properties.multi
    );
  }, [properties.default, properties.options, data, stateValue]);

  // Bind the selected option to the reducer state
  useEffect(() => {
    // If we haven't got the data we need yet...
    if (!options) {
      return;
    }

    // If this is first load and we have a value from state, initialise it
    if (!initialisedFromState && stateValue) {
      const parsedUrlValue = properties.multi
        ? stateValue.split(",")
        : stateValue;
      const foundOptions = findOptions(
        options,
        properties.multi,
        parsedUrlValue
      );
      setValue(foundOptions || null);
      setInitialisedFromState(true);
    } else if (!initialisedFromState && properties.default) {
      const parsedUrlValue = properties.multi
        ? properties.default.split(",")
        : properties.default;
      const foundOptions = findOptions(
        options,
        properties.multi,
        parsedUrlValue
      );
      setValue(foundOptions || null);
      setInitialisedFromState(true);
    } else if (
      !initialisedFromState &&
      !stateValue &&
      !properties.default &&
      properties.placeholder
    ) {
      setInitialisedFromState(true);
    } else if (
      !initialisedFromState &&
      !stateValue &&
      !properties.default &&
      !properties.placeholder
    ) {
      setInitialisedFromState(true);
      setValue(options[0]);
      dispatch({
        type: DashboardActions.SET_DASHBOARD_INPUT,
        name,
        value: getValueForState(properties.multi, options[0]),
        recordInputsHistory: false,
      });
    } else {
      if (
        initialisedFromState &&
        stateValue &&
        value &&
        // @ts-ignore
        stateValue !== value.value
      ) {
        const parsedUrlValue = properties.multi
          ? stateValue.split(",")
          : stateValue;
        const foundOptions = findOptions(
          options,
          properties.multi,
          parsedUrlValue
        );
        setValue(foundOptions || null);
      } else if (initialisedFromState && !stateValue && !properties.default) {
        setValue(null);
      }
    }
  }, [
    dispatch,
    initialisedFromState,
    name,
    options,
    properties.default,
    properties.multi,
    properties.placeholder,
    stateValue,
    value,
  ]);

  useEffect(() => {
    if (!initialisedFromState) {
      return;
    }

    // @ts-ignore
    if (!value || value.length === 0) {
      dispatch({
        type: DashboardActions.DELETE_DASHBOARD_INPUT,
        name,
        recordInputsHistory: true,
      });
      return;
    }

    dispatch({
      type: DashboardActions.SET_DASHBOARD_INPUT,
      name,
      value: getValueForState(properties.multi, value),
      recordInputsHistory: true,
    });
  }, [dispatch, initialisedFromState, properties.multi, name, value]);

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
        isClearable={!!properties.placeholder}
        menuIsOpen={true}
        isRtl={false}
        isSearchable
        isMulti={properties.multi}
        // menuIsOpen
        name={name}
        // @ts-ignore
        onChange={(value) => setValue(value)}
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
