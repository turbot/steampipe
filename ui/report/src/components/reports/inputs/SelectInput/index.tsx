import Icon from "../../../Icon";
import { classNames } from "../../../../utils/styles";
import { Fragment, useState } from "react";
import { IInput, InputProps } from "../index";
import { Listbox, Transition } from "@headlessui/react";
import {
  openSelectMenuIcon,
  selectMenuItemSelectedIcon,
} from "../../../../constants/icons";

interface SelectInputItem {
  label: string;
  value: any;
}

const SelectInput = (props: InputProps) => {
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

  return (
    <div>
      <label
        htmlFor={props.name}
        className="block text-sm font-medium text-gray-700"
      >
        {props.title}
      </label>
      <select
        id={props.name}
        name="location"
        className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-black-scale-3 bg-background focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
        // defaultValue="Canada"
      >
        <option value={undefined}>Please select...</option>
        {(props.data ? props.data.items || [] : []).map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    </div>
  );
};

const definition: IInput = {
  type: "select",
  component: SelectInput,
};

export default definition;
