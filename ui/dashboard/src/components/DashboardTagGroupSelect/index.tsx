import { useDashboard } from "../../hooks/useDashboard";
import { Fragment, useEffect, useMemo, useState } from "react";
import { useParams, useSearchParams } from "react-router-dom";
import { Listbox, Transition } from "@headlessui/react";
import { CheckIcon, SelectorIcon } from "@heroicons/react/solid";
import { classNames } from "../../utils/styles";
import { startCase } from "lodash";

const DashboardTagGroupSelect = () => {
  const { dashboardSearch, dashboardTagKeys } = useDashboard();
  const { dashboardName } = useParams();
  const [searchParams, setSearchParams] = useSearchParams();
  const [value, setValue] = useState({
    groupBy: "tag",
    tag: "service",
    label: "Service",
  });
  // const [groupBy, setGroupBy] = useState(searchParams.get("group_by") || "tag");
  // const [tag, setTag] = useState(searchParams.get("tag") || "service");

  useEffect(() => {
    searchParams.set("group_by", value.groupBy);
    if (value.tag) {
      searchParams.set("tag", value.tag);
    } else {
      searchParams.delete("tag");
    }
    setSearchParams(searchParams);
  }, [searchParams, value]);

  useEffect(() => {
    if (dashboardName && !dashboardSearch) {
      searchParams.delete("group_by");
      searchParams.delete("tag");
      setSearchParams(searchParams);
    }
  }, [dashboardName, dashboardSearch, searchParams]);

  const options = useMemo(() => {
    const o = [{ groupBy: "mod", tag: "", label: "Mod" }];
    for (const dashboardTagKey of dashboardTagKeys) {
      o.push({
        groupBy: "tag",
        tag: dashboardTagKey,
        label: startCase(dashboardTagKey),
      });
    }
    return o;
  }, [dashboardTagKeys]);

  // const value = useMemo(() => ({ groupBy, tag }), [groupBy, tag]);

  // const updateValues = (selectedValue) => {
  //   setGroupBy(selectedValue.groupBy);
  //   setTag(selectedValue.tag);
  // };

  return (
    <Listbox value={value} onChange={setValue}>
      {({ open }) => (
        <>
          <div className="relative">
            <Listbox.Button className="relative w-full bg-white border border-gray-300 rounded-md shadow-sm pl-3 pr-10 py-2 text-left cursor-default focus:outline-none focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 ">
              <span className="block truncate">Group by: {value.label}</span>
              <span className="absolute inset-y-0 right-0 flex items-center pr-2 pointer-events-none">
                <SelectorIcon
                  className="h-5 w-5 text-gray-400"
                  aria-hidden="true"
                />
              </span>
            </Listbox.Button>

            <Transition
              show={open}
              as={Fragment}
              leave="transition ease-in duration-100"
              leaveFrom="opacity-100"
              leaveTo="opacity-0"
            >
              <Listbox.Options className="absolute z-10 mt-1 w-full bg-white shadow-lg max-h-60 rounded-md py-1 text-base ring-1 ring-black ring-opacity-5 overflow-auto focus:outline-none sm:text-sm">
                {options.map((option) => (
                  <Listbox.Option
                    key={`${option.groupBy}:${option.tag}`}
                    // @ts-ignore
                    className={({ active }) =>
                      classNames(
                        active ? "text-white bg-indigo-600" : "text-gray-900",
                        "cursor-default select-none relative py-2 pl-8 pr-4"
                      )
                    }
                    value={option}
                  >
                    {({ selected, active }) => (
                      <>
                        <span
                          className={classNames(
                            selected ? "font-semibold" : "font-normal",
                            "block truncate"
                          )}
                        >
                          {option.label}
                        </span>

                        {selected ? (
                          <span
                            className={classNames(
                              active ? "text-white" : "text-indigo-600",
                              "absolute inset-y-0 left-0 flex items-center pl-1.5"
                            )}
                          >
                            <CheckIcon className="h-5 w-5" aria-hidden="true" />
                          </span>
                        ) : null}
                      </>
                    )}
                  </Listbox.Option>
                ))}
              </Listbox.Options>
            </Transition>
          </div>
        </>
      )}
    </Listbox>
  );
};

export default DashboardTagGroupSelect;
