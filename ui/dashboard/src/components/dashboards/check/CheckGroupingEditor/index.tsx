import Icon from "../../../Icon";
import Select from "react-select";
import useDeepCompareEffect from "use-deep-compare-effect";
import useSelectInputStyles from "../../inputs/common/useSelectInputStyles";
import { CheckDisplayGroup, CheckDisplayGroupType } from "../common";
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

type CheckGroupingEditorProps = {
  config: CheckDisplayGroup[];
  setConfig: (newValue: CheckDisplayGroup[]) => void;
};

type CheckGroupingEditorItemProps = {
  config: CheckDisplayGroup[];
  item: CheckDisplayGroup;
  index: number;
  remove: (index: number) => void;
  update: (index: number, item: CheckDisplayGroup) => void;
};

type CheckGroupingTypeSelectProps = {
  config: CheckDisplayGroup[];
  index: number;
  item: CheckDisplayGroup;
  type: CheckDisplayGroupType;
  update: (index: number, updatedItem: CheckDisplayGroup) => void;
};

type CheckGroupingValueSelectProps = {
  index: number;
  item: CheckDisplayGroup;
  type: CheckDisplayGroupType;
  update: (index: number, updatedItem: CheckDisplayGroup) => void;
  value: string | undefined;
};

const CheckGroupingTypeSelect = ({
  config,
  index,
  item,
  type,
  update,
}: CheckGroupingTypeSelectProps) => {
  const [currentType, setCurrentType] = useState<CheckDisplayGroupType>(type);

  useDeepCompareEffect(() => {
    console.log("Setting CheckGroupingTypeSelect", {
      currentType,
      index,
      item,
      update,
    });

    update(index, {
      ...item,
      id: currentType,
      type: currentType,
      value: "",
    });
  }, [currentType, index, item]);

  const types = useMemo(() => {
    const existingTypes = config.map((c) => c.type.toString());
    const allTypes: SelectOption[] = [
      { value: "benchmark", label: "Benchmark" },
      { value: "control", label: "Control" },
      { value: "dimension", label: "Dimension" },
      { value: "reason", label: "Reason" },
      { value: "resource", label: "Resource" },
      { value: "result", label: "Result" },
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
      placeholder="Select a grouping…"
      styles={styles}
      value={types.find((t) => t.value === type)}
    />
  );
};

const CheckGroupingValueSelect = ({
  index,
  item,
  type,
  value,
  update,
}: CheckGroupingValueSelectProps) => {
  const [currentValue, setCurrentValue] = useState(value);
  const { context: filterValues } = useDashboardControls();

  useDeepCompareEffect(() => {
    console.log("Setting CheckGroupingValueSelect", {
      currentValue,
      index,
      item,
      update,
    });
    update(index, {
      ...item,
      id: `${item.type}-${currentValue}`,
      value: currentValue,
    });
  }, [currentValue, index, item]);

  const values = useMemo(() => {
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
      onChange={(t) => setCurrentValue((t as SelectOption).value)}
      options={values}
      inputId={`${type}.input`}
      placeholder="Select a grouping…"
      styles={styles}
      value={values.find((t) => t.value === value)}
    />
  );
};

const CheckGroupingEditorItem = ({
  config,
  index,
  item,
  remove,
  update,
}: CheckGroupingEditorItemProps) => {
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
        <CheckGroupingTypeSelect
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
            <CheckGroupingValueSelect
              index={index}
              item={item}
              type={item.type}
              update={update}
              value={item.value}
            />
          </div>
        </>
      )}
      <span
        className={classNames(
          config.length > 1
            ? "text-foreground-light hover:text-steampipe-red cursor-pointer"
            : "text-foreground-lightest",
        )}
        onClick={config.length > 1 ? () => remove(index) : undefined}
        title={
          config.length > 1
            ? "Remove"
            : "Grouping must contain at least one level"
        }
      >
        <Icon className="h-5 w-5" icon="trash" />
      </span>
    </Reorder.Item>
  );
};

const CheckGroupingEditor = ({
  config,
  setConfig,
}: CheckGroupingEditorProps) => {
  const remove = useCallback(
    (index: number) => {
      const removed = [...config.slice(0, index), ...config.slice(index + 1)];
      setConfig(removed);
    },
    [config, setConfig],
  );

  const update = useCallback(
    (index: number, updatedItem: CheckDisplayGroup) => {
      const updated = [
        ...config.slice(0, index),
        updatedItem,
        ...config.slice(index + 1),
      ];
      setConfig(updated);
    },
    [config, setConfig],
  );

  return (
    <div className="flex flex-col space-y-4">
      <Reorder.Group
        axis="y"
        values={config}
        onReorder={setConfig}
        as="div"
        className="flex flex-col space-y-4"
      >
        {config.map((c, idx) => (
          <CheckGroupingEditorItem
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
        onClick={() => setConfig([...config, { id: "", type: "" }])}
      >
        <Icon className="inline-block h-4 w-4" icon="plus" />
        <span className="inline-block">Add grouping</span>
      </div>
    </div>
  );
};

export default CheckGroupingEditor;
