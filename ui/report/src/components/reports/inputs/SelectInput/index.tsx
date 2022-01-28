import Select from "react-select";
import { getColumnIndex } from "../../../../utils/data";
import { IInput, InputProps } from "../index";
import { ThemeNames, useTheme } from "../../../../hooks/useTheme";
import { useEffect, useMemo, useState } from "react";

const SelectInput = (props: InputProps) => {
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

  // const [selected, setSelected] = useState<SelectInputItem | null>(null);
  // console.log(props.data);

  // return (
  //   <Listbox value={selected} onChange={setSelected}>
  //     {({ open }) => {
  //       console.log(open);
  //       return (
  //         <>
  //           <Listbox.Label className="block text-sm font-medium text-gray-700">
  //             {props.title}
  //           </Listbox.Label>
  //           <div className="mt-1 relative">
  //             <Listbox.Button className="relative w-full bg-white border border-gray-300 rounded-md shadow-sm pl-3 pr-10 py-2 text-left cursor-default focus:outline-none focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm">
  //               {
  //                 <span className="block truncate">
  //                   {selected ? selected.label : "Please select..."}
  //                 </span>
  //               }
  //               <span className="absolute inset-y-0 right-0 flex items-center pr-2 pointer-events-none">
  //                 <Icon
  //                   className="h-5 w-5 text-gray-400"
  //                   aria-hidden="true"
  //                   icon={openSelectMenuIcon}
  //                 />
  //               </span>
  //             </Listbox.Button>
  //
  //             <Transition
  //               show={open}
  //               as={Fragment}
  //               leave="transition ease-in duration-100"
  //               leaveFrom="opacity-100"
  //               leaveTo="opacity-0"
  //             >
  //               <Listbox.Options className="absolute z-50 w-full h-full bg-white shadow-lg max-h-60 rounded-md py-1 text-base ring-1 ring-black ring-opacity-5 overflow-auto focus:outline-none sm:text-sm">
  //                 {(props.data ? props.data.items || [] : []).map((option) => {
  //                   console.log(option);
  //                   return (
  //                     <Listbox.Option
  //                       key={option.value}
  //                       className={({ active }) =>
  //                         classNames(
  //                           active
  //                             ? "text-white bg-indigo-600"
  //                             : "text-gray-900",
  //                           "cursor-default select-none relative py-2 pl-8 pr-4"
  //                         )
  //                       }
  //                       value={option}
  //                     >
  //                       {({ selected, active }) => (
  //                         <>
  //                           <span
  //                             className={classNames(
  //                               selected ? "font-semibold" : "font-normal",
  //                               "block truncate"
  //                             )}
  //                           >
  //                             {option.label}
  //                           </span>
  //
  //                           {selected ? (
  //                             <span
  //                               className={classNames(
  //                                 active ? "text-white" : "text-indigo-600",
  //                                 "absolute inset-y-0 left-0 flex items-center pl-1.5"
  //                               )}
  //                             >
  //                               <Icon
  //                                 className="h-5 w-5"
  //                                 aria-hidden="true"
  //                                 icon={selectMenuItemSelectedIcon}
  //                               />
  //                             </span>
  //                           ) : null}
  //                         </>
  //                       )}
  //                     </Listbox.Option>
  //                   );
  //                 })}
  //               </Listbox.Options>
  //             </Transition>
  //           </div>
  //         </>
  //       );
  //     }}
  //   </Listbox>
  // );

  // return (
  //   <div>
  //     <label
  //       htmlFor={props.name}
  //       className="block text-sm font-medium text-gray-700"
  //     >
  //       {props.title}
  //     </label>
  //     <select
  //       id={props.name}
  //       name="location"
  //       className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-black-scale-3 bg-background focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
  //       // defaultValue="Canada"
  //     >
  //       <option value={undefined}>Please select...</option>
  //       {(props.data ? props.data.items || [] : []).map((option) => (
  //         <option key={option.value} value={option.value}>
  //           {option.label}
  //         </option>
  //       ))}
  //     </select>
  //   </div>
  // );
};

const definition: IInput = {
  type: "select",
  component: SelectInput,
};

export default definition;
