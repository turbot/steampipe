import { CheckIcon, SelectorIcon } from "@heroicons/react/solid";
import { classNames } from "../../utils/styles";
import { Fragment, useEffect, useMemo, useState } from "react";
import { Listbox, Transition } from "@headlessui/react";
import { sortBy, startCase } from "lodash";
import { useDashboard } from "../../hooks/useDashboard";
import { useParams, useSearchParams } from "react-router-dom";

const DashboardTagGroupSelect = () => {
  const { dashboardSearch, dashboardTagKeys, availableDashboardsLoaded } =
    useDashboard();
  const { dashboardName } = useParams();
  const [searchParams, setSearchParams] = useSearchParams();
  // const [groupBy, setGroupBy] = useState(searchParams.get("group_by") || "tag");
  // const [tag, setTag] = useState(searchParams.get("tag") || "service");

  const options = useMemo(() => {
    const o = [
      { groupBy: "mod", tag: "", label: "Mod" },
      {
        groupBy: "tag",
        tag: "category",
        label: "Category",
      },
      {
        groupBy: "tag",
        tag: "service",
        label: "Service",
      },
      {
        groupBy: "tag",
        tag: "type",
        label: "Type",
      },
    ];
    for (const dashboardTagKey of dashboardTagKeys) {
      if (!o.find((i) => i.tag === dashboardTagKey)) {
        o.push({
          groupBy: "tag",
          tag: dashboardTagKey,
          label: startCase(dashboardTagKey),
        });
      }
    }
    return sortBy(o, ["label"]);
  }, [dashboardTagKeys]);

  const [value, setValue] = useState(() => {
    let option = options.find((o) => o.tag === "service");
    if (!option) {
      option = options.find((o) => o.groupBy === "mod");
    }
    return option;
  });

  useEffect(() => {
    if (!value) {
      return;
    }
    // @ts-ignore
    searchParams.set("group_by", value.groupBy);
    // @ts-ignore
    if (value.tag) {
      // @ts-ignore
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

  // useEffect(() => {
  //   console.log(availableDashboardsLoaded, dashboardTagKeys, options);
  //   if (availableDashboardsLoaded) {
  //     return;
  //   }
  //   let option;
  //   const hasServiceTag = dashboardTagKeys.includes("service");
  //   if (hasServiceTag) {
  //     option = options.find((o) => o.tag === "service");
  //   } else {
  //     option = options.find((o) => o.groupBy === "mod");
  //   }
  //   console.log("Setting", option);
  //   setValue(option);
  // }, [availableDashboardsLoaded, dashboardTagKeys, options]);

  // const value = useMemo(() => ({ groupBy, tag }), [groupBy, tag]);

  // const updateValues = (selectedValue) => {
  //   setGroupBy(selectedValue.groupBy);
  //   setTag(selectedValue.tag);
  // };

  if (!availableDashboardsLoaded || !value) {
    return null;
  }

  return (
    <Listbox value={value} onChange={setValue}>
      {({ open }) => (
        <>
          <div className="relative">
            <Listbox.Button className="relative w-full bg-background border border-table-border rounded-md shadow-sm pl-3 pr-10 py-2 text-left text-sm md:text-base cursor-default focus:outline-none focus:ring-1">
              {/*@ts-ignore*/}
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
              <Listbox.Options className="absolute z-10 mt-1 w-full bg-background shadow-lg max-h-60 rounded-md py-1 text-base ring-1 ring-black ring-opacity-5 overflow-auto focus:outline-none sm:text-sm">
                {options.map((option) => (
                  <Listbox.Option
                    key={`${option.groupBy}:${option.tag}`}
                    // @ts-ignore
                    className={({ active }) =>
                      classNames(
                        active
                          ? "text-foreground bg-black-scale-1"
                          : "text-foreground",
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
                            className={
                              "absolute inset-y-0 left-0 flex items-center pl-1.5"
                            }
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
