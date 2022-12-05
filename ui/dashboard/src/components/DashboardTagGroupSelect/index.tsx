import sortBy from "lodash/sortBy";
import { CheckIcon, ChevronUpDownIcon } from "@heroicons/react/24/solid";
import { classNames } from "../../utils/styles";
import { DashboardActions } from "../../types";
import { Fragment, useCallback, useEffect, useMemo, useState } from "react";
import { Listbox, Transition } from "@headlessui/react";
import { useDashboard } from "../../hooks/useDashboard";
import { useParams } from "react-router-dom";

const DashboardTagGroupSelect = () => {
  const { availableDashboardsLoaded, dispatch, search } = useDashboard();
  const { dashboard_name } = useParams();

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
    // for (const dashboardTagKey of dashboardTags.keys) {
    //   if (!o.find((i) => i.tag === dashboardTagKey)) {
    //     o.push({
    //       groupBy: "tag",
    //       tag: dashboardTagKey,
    //       label: startCase(dashboardTagKey),
    //     });
    //   }
    // }
    return sortBy(o, ["label"]);
  }, []);

  const findOption = useCallback(
    (groupBy) => {
      if (groupBy.value === "tag") {
        return options.find((o) => o.tag === groupBy.tag);
      }
      return options.find((o) => o.groupBy === "mod");
    },
    [options]
  );

  const [value, setValue] = useState(() => findOption(search.groupBy));

  const updateState = useCallback(
    (option) =>
      dispatch({
        type: DashboardActions.SET_DASHBOARD_SEARCH_GROUP_BY,
        value: option.groupBy,
        tag: option.tag,
      }),
    [dispatch]
  );

  useEffect(() => {
    setValue(findOption(search.groupBy));
  }, [findOption, search.groupBy]);

  if (
    !availableDashboardsLoaded ||
    !value ||
    (dashboard_name && !search.value)
  ) {
    return null;
  }

  return (
    <Listbox value={value} onChange={updateState}>
      {({ open }) => (
        <>
          <div className="relative">
            <Listbox.Button className="relative w-full bg-dashboard-panel border border-table-border rounded-md pl-3 pr-7 md:pr-10 py-2 text-left text-sm md:text-base cursor-pointer focus:ring-1 focus:ring-text-link">
              {/*@ts-ignore*/}
              <span className="block truncate">
                <span className="hidden md:inline mr-1">Group by:</span>
                {value.label}
              </span>
              <span className="absolute inset-y-0 right-0 flex items-center pr-1 md:pr-2 pointer-events-none">
                <ChevronUpDownIcon
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
              <Listbox.Options className="absolute z-10 w-32 sm:w-full bg-dashboard-panel shadow-lg max-h-60 rounded-md text-base ring-1 ring-black ring-opacity-5 overflow-auto focus:outline-none sm:text-sm">
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
                    {({ selected }) => (
                      <>
                        <span className="block truncate">{option.label}</span>
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
