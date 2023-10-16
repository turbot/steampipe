import Icon from "../../../Icon";
import Modal from "../../../Modal";
import NeutralButton from "../../../forms/NeutralButton";
import StartCase from "lodash/startCase";
import { classNames } from "../../../../utils/styles";
import { CheckDisplayGroup, CheckDisplayGroupType } from "../common";
import { Reorder, useDragControls } from "framer-motion";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useSearchParams } from "react-router-dom";

type CheckGroupingEditorModalProps = {
  config: CheckDisplayGroup[];
  setShowEditor: (show: boolean) => void;
};

type CheckGroupingEditorProps = {
  config: CheckDisplayGroup[];
  setCurrentConfig: (newConfig: CheckDisplayGroup[]) => void;
  setIsValid: (value: boolean) => void;
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
  type: CheckDisplayGroupType;
  update: (type: CheckDisplayGroupType) => void;
};

const CheckGroupingTypeSelect = ({
  config,
  type,
  update,
}: CheckGroupingTypeSelectProps) => {
  const [currentType, setCurrentType] = useState(type);

  useEffect(() => {
    update(currentType);
  }, [currentType]);

  const types = useMemo(() => {
    const existingTypes = config.map((c) => c.type);
    const allTypes: CheckDisplayGroupType[] = [
      "benchmark",
      "control",
      "dimension",
      "reason",
      "resource",
      "result",
      "severity",
      "status",
      "tag",
    ];
    return allTypes.filter(
      (t) => t === type || t === "dimension" || !existingTypes.includes(t),
    );
  }, [config, type]);

  return (
    <select
      id="grouping-type"
      className="block w-full rounded-md bg-dashboard border-black-scale-2 py-2 pl-3 pr-10 text-foreground focus:border-indigo-500 focus:outline-none focus:ring-indigo-500"
      name="grouping-type"
      onChange={(e) => {
        setCurrentType(e.target.value as CheckDisplayGroupType);
      }}
      value={type}
    >
      <option defaultValue={currentType} value={""}>
        Select a grouping type...
      </option>
      {types.map((t) => (
        <option key={t} className="capitalize" value={t}>
          {StartCase(t)}
        </option>
      ))}
    </select>
  );
  // return (
  //   <Listbox value={currentType} onChange={setCurrentType}>
  //     <Listbox.Button>
  //       <span className="capitalize">{currentType}</span>{" "}
  //     </Listbox.Button>
  //     <Listbox.Options>
  //       {types.map((t) => (
  //         <Listbox.Option key={t} value={t}>
  //           <span className="capitalize">{t}</span>
  //         </Listbox.Option>
  //       ))}
  //     </Listbox.Options>
  //   </Listbox>
  // <select>
  //   {types.map((t) => (
  //     <option className="capitalize" selected={currentType === t} value={t}>
  //       <span className="capitalize">{t}</span>{" "}
  //     </option>
  //   ))}
  // </select>
  // );
};

const CheckGroupingEditorItem = ({
  config,
  index,
  item,
  remove,
  update,
}: CheckGroupingEditorItemProps) => {
  const dragControls = useDragControls();

  useEffect(() => {
    if (item.type !== "dimension" && item.type !== "tag" && item.value) {
      update(index, { ...item, id: item.type, value: undefined });
    } else if (
      (item.type === "dimension" || item.type === "tag") &&
      item.value
    ) {
      update(index, { ...item, id: `${item.type}-${item.value}` });
    }
  }, [item.type, item.value, update]);

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
        <Icon className="h-5 w-5" icon="drag_handle" />
      </div>
      <div className="grow">
        <CheckGroupingTypeSelect
          type={item.type}
          config={config}
          update={(updated) =>
            update(index, {
              ...item,
              id: updated,
              value: "",
              type: updated,
            })
          }
        />
      </div>
      {item.type === "dimension" && (
        <>
          <span>=</span>
          <div className="grow">
            <input
              className="flex w-full p-2 bg-dashboard text-foreground border border-divide rounded-md"
              onChange={(e) =>
                update(index, {
                  ...item,
                  id: `${item.type}-${e.target.value}`,
                  value: e.target.value,
                })
              }
              value={item.value || ""}
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
        title="Remove"
      >
        <Icon className="h-5 w-5" icon="trash" />
      </span>
    </Reorder.Item>
  );
};

const CheckGroupingEditor = ({
  config,
  setCurrentConfig,
  setIsValid,
}: CheckGroupingEditorProps) => {
  useEffect(() => {
    const isValid = config.every((c) => {
      switch (c.type) {
        case "benchmark":
        case "control":
        case "result":
        case "reason":
        case "resource":
        case "severity":
        case "status":
          return !c.value;
        case "dimension":
        case "tag":
          return !!c.value;
      }
    });
    setIsValid(isValid);
  }, [config, setIsValid]);

  const remove = useCallback(
    (index: number) => {
      const removed = [...config.slice(0, index), ...config.slice(index + 1)];
      setCurrentConfig(removed);
    },
    [config, setCurrentConfig],
  );

  const update = useCallback(
    (index: number, item: CheckDisplayGroup) => {
      const updated = [
        ...config.slice(0, index),
        item,
        ...config.slice(index + 1),
      ];
      setCurrentConfig(updated);
    },
    [config, setCurrentConfig],
  );

  // const swap = useCallback(
  //   (x: number, y: number) => {
  //     console.log(x, y);
  //     const newConfig = [...config];
  //     const temp = newConfig[x];
  //     newConfig[x] = newConfig[y];
  //     newConfig[y] = temp;
  //     setInnerConfig(newConfig);
  //   },
  //   [config, setInnerConfig],
  // );

  return (
    <div className="flex flex-col space-y-4">
      <Reorder.Group
        axis="y"
        values={config}
        onReorder={setCurrentConfig}
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
        className="flex items-center text-link cursor-pointer space-x-1"
        // @ts-ignore
        onClick={() => setCurrentConfig([...config, { id: "", type: "" }])}
      >
        <Icon className="inline-block h-4 w-4" icon="plus" />
        <span className="inline-block">Add grouping</span>
      </div>
    </div>
  );
};

const CheckGroupingEditorModal = ({
  config,
  setShowEditor,
}: CheckGroupingEditorModalProps) => {
  const [currentConfig, setCurrentConfig] = useState(config);
  const [isValid, setIsValid] = useState(false);
  const [_, setSearchParams] = useSearchParams();

  const saveGroupingConfig = () => {
    setSearchParams({
      grouping: currentConfig
        .map((c) =>
          c.type === "dimension" || c.type === "tag"
            ? `${c.type}|${c.value}`
            : c.type,
        )
        .join(","),
    });
  };

  return (
    <Modal
      actions={[
        <NeutralButton key="cancel" onClick={() => setShowEditor(false)}>
          <>Cancel</>
        </NeutralButton>,
        <NeutralButton
          key="save"
          disabled={currentConfig.length === 0 || !isValid}
          onClick={() => {
            saveGroupingConfig();
            setShowEditor(false);
          }}
        >
          <>Save</>
        </NeutralButton>,
      ]}
      icon={<Icon className="h-6 w-6" icon="pencil-square" />}
      onClose={async () => {
        setShowEditor(false);
      }}
      title="Edit benchmark grouping"
    >
      <CheckGroupingEditor
        config={currentConfig}
        setCurrentConfig={setCurrentConfig}
        setIsValid={setIsValid}
      />
    </Modal>
  );
};

export { CheckGroupingEditorModal };
