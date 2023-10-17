import CreatableSelect from "react-select/creatable";
import Icon from "../../../Icon";
import Select from "react-select";
import useDeepCompareEffect from "use-deep-compare-effect";
import useSelectInputStyles from "../../inputs/common/useSelectInputStyles";
import { CheckFilter, CheckFilterType, Filter } from "../common";
import { classNames } from "../../../../utils/styles";
import {
  MultiValueLabelWithTags,
  OptionWithTags,
  SingleValueWithTags,
} from "../../inputs/common/Common";
import { Reorder, useDragControls } from "framer-motion";
import { SelectOption } from "../../inputs/types";
import { useCallback, useMemo, useState } from "react";
import { useDashboardControls } from "../../layout/Dashboard/DashboardControlsProvider";

type CheckFilterEditorProps = {
  config: CheckFilter;
  setConfig: (newValue: CheckFilter) => void;
};

type CheckFilterEditorItemProps = {
  config: CheckFilter;
  item: Filter;
  index: number;
  remove: (index: number) => void;
  update: (index: number, item: Filter) => void;
};

type CheckFilterTypeSelectProps = {
  config: CheckFilter;
  index: number;
  item: Filter;
  type: CheckFilterType;
  update: (index: number, updatedItem: Filter) => void;
};

type CheckFilterKeySelectProps = {
  index: number;
  item: Filter;
  type: CheckFilterType;
  update: (index: number, updatedItem: Filter) => void;
  filterKey: string | undefined;
};

type CheckFilterValueSelectProps = {
  index: number;
  item: Filter;
  type: CheckFilterType;
  update: (index: number, updatedItem: Filter) => void;
  value: string | undefined;
};

const CheckFilterTypeSelect = ({
  config,
  index,
  item,
  type,
  update,
}: CheckFilterTypeSelectProps) => {
  const [currentType, setCurrentType] = useState<CheckFilterType>(type);

  useDeepCompareEffect(() => {
    console.log("Setting CheckFilterTypeSelect", {
      currentType,
      index,
      item,
      update,
    });

    update(index, {
      ...item,
      type: currentType,
      value: "",
    });
  }, [currentType, index, item]);

  const types = useMemo(() => {
    // @ts-ignore
    const existingTypes = config.and.map((c) => c.type.toString());
    const allTypes: SelectOption[] = [
      { value: "benchmark", label: "Benchmark" },
      { value: "control", label: "Control" },
      { value: "dimension", label: "Dimension" },
      { value: "reason", label: "Reason" },
      { value: "resource", label: "Resource" },
      { value: "severity", label: "Severity" },
      { value: "status", label: "Status" },
      { value: "tag", label: "Tag" },
    ];
    return allTypes.filter(
      (t) =>
        t.value === type ||
        t.value === "dimension" ||
        t.value === "tag" ||
        // @ts-ignore
        !existingTypes.includes(t.value),
    );
  }, [config, type]);

  const styles = useSelectInputStyles();

  return (
    <Select
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
      // @ts-ignore as this element definitely exists
      menuPortalTarget={document.getElementById("portals")}
      onChange={(t) => setCurrentType((t as SelectOption).value)}
      options={types}
      inputId={`${type}.input`}
      placeholder="Select a filter…"
      styles={styles}
      value={types.find((t) => t.value === type)}
    />
  );
};

const CheckFilterKeySelect = ({
  index,
  item,
  type,
  filterKey,
  update,
}: CheckFilterKeySelectProps) => {
  const [currentKey, setCurrentKey] = useState(filterKey);
  const { context: filterValues } = useDashboardControls();

  useDeepCompareEffect(() => {
    console.log("Setting CheckFilterKeySelect", {
      currentKey,
      index,
      item,
      update,
    });
    update(index, {
      ...item,
      key: currentKey,
    });
  }, [currentKey, index, item]);

  const keys = useMemo(() => {
    return Object.keys(filterValues[type].key || {}).map((k) => ({
      value: k,
      label: k,
    }));
  }, [filterValues, type]);

  const styles = useSelectInputStyles();

  return (
    <Select
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
      // @ts-ignore as this element definitely exists
      menuPortalTarget={document.getElementById("portals")}
      onChange={(t) => setCurrentKey((t as SelectOption).value)}
      options={keys}
      inputId={`${type}.input`}
      placeholder="Enter a filter…"
      styles={styles}
      value={keys.find((t) => t.value === filterKey)}
    />
  );
};

const CheckFilterValueSelect = ({
  index,
  item,
  type,
  value,
  update,
}: CheckFilterValueSelectProps) => {
  const [currentValue, setCurrentValue] = useState(value);
  const { context: filterValues } = useDashboardControls();

  useDeepCompareEffect(() => {
    console.log("Setting CheckFilterValueSelect", {
      currentValue,
      index,
      item,
      update,
    });
    update(index, {
      ...item,
      value: currentValue,
    });
  }, [currentValue, index, item]);

  const values = useMemo(() => {
    console.log(filterValues);
    if (!type) {
      return [];
    }
    if (type === "status") {
      return (
        Object.entries(filterValues[type] || {})
          // @ts-ignore
          .filter(([k, v]) => v > 0)
          .map(([k]) => ({
            value: k,
            label: k,
          }))
      );
    }
    return Object.keys(filterValues[type].value || {}).map((k) => ({
      value: k,
      label: k,
    }));
  }, [filterValues, type]);

  const styles = useSelectInputStyles();

  return (
    <CreatableSelect
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
      onChange={(t) => setCurrentValue((t as SelectOption).value)}
      options={values}
      inputId={`${type}.input`}
      placeholder="Enter a filter…"
      styles={styles}
      value={values.find((t) => t.value === value)}
    />
  );
};

const CheckFilterEditorItem = ({
  config,
  index,
  item,
  remove,
  update,
}: CheckFilterEditorItemProps) => {
  const dragControls = useDragControls();

  return (
    <Reorder.Item
      as="div"
      id={`${item.type}-${item.value}`}
      className="flex space-x-3 items-center"
      dragControls={dragControls}
      dragListener={false}
      value={item}
    >
      {/*<div className="flex space-x-3 items-center">*/}
      <div className="cursor-grab" onPointerDown={(e) => dragControls.start(e)}>
        <Icon className="h-5 w-5" icon="drag_indicator" />
      </div>
      <div className="grow">
        <CheckFilterTypeSelect
          config={config}
          index={index}
          item={item}
          type={item.type}
          update={update}
        />
      </div>
      {(item.type === "dimension" || item.type === "tag") && (
        <>
          <span>=</span>
          <div className="grow">
            <CheckFilterKeySelect
              index={index}
              item={item}
              filterKey={item.key}
              type={item.type}
              update={update}
            />
          </div>
        </>
      )}
      <span>=</span>
      <div className="grow">
        <CheckFilterValueSelect
          index={index}
          item={item}
          type={item.type}
          update={update}
          value={item.value}
        />
      </div>
      <span
        className={classNames(
          // @ts-ignore
          config.and.length > 1
            ? "text-foreground-light hover:text-steampipe-red cursor-pointer"
            : "text-foreground-lightest",
        )}
        // @ts-ignore
        onClick={() => remove(index)}
        title="Remove"
      >
        <Icon className="h-5 w-5" icon="trash" />
      </span>
    </Reorder.Item>
  );
};

const CheckFilterEditor = ({ config, setConfig }: CheckFilterEditorProps) => {
  const remove = useCallback(
    (index: number) => {
      const newConfig: CheckFilter = {
        and: config?.and || [],
      };
      if (newConfig.and) {
        newConfig.and = [
          ...newConfig.and.slice(0, index),
          ...newConfig.and.slice(index + 1),
        ];
      }
      setConfig(newConfig);
    },
    [config, setConfig],
  );

  const update = useCallback(
    (index: number, updatedItem: Filter) => {
      const newConfig: CheckFilter = {
        and: config?.and || [],
      };
      if (newConfig.and) {
        newConfig.and[index] = updatedItem;
      }
      setConfig(newConfig);
    },
    [config, setConfig],
  );

  return (
    <div className="flex flex-col space-y-4">
      <Reorder.Group
        axis="y"
        values={config?.and || []}
        onReorder={(a) => {
          if (!!config) {
            const newConfig = {
              ...config,
              and: a,
            };
            setConfig(newConfig);
          }
        }}
        as="div"
        className="flex flex-col space-y-4"
      >
        {/*@ts-ignore*/}
        {(config?.and || []).map((c: Filter, idx: number) => (
          <CheckFilterEditorItem
            key={`${c.type}-${c.value}`}
            config={config}
            item={c}
            index={idx}
            remove={remove}
            update={update}
          />
        ))}
      </Reorder.Group>
      <div
        className="flex items-center text-link cursor-pointer space-x-3"
        // @ts-ignore
        onClick={() => setConfig({ and: [...config.and, { type: "" }] })}
      >
        <Icon className="inline-block h-4 w-4" icon="plus" />
        <span className="inline-block">Add filter</span>
      </div>
    </div>
  );
};

export default CheckFilterEditor;
