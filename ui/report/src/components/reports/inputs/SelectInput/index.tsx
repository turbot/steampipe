import Select from "react-select";
import { IInput, InputProps } from "../index";
import { useMemo } from "react";
import { getColumnIndex } from "../../../../utils/data";

const SelectInput = (props: InputProps) => {
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

  return (
    <form>
      {props.title && (
        <label
          className="mb-2 text-sm"
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
        name={props.name}
        options={options}
        placeholder={
          (props.properties && props.properties.placeholder) ||
          "Please select..."
        }
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
