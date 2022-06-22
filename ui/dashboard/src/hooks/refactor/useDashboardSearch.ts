import {
  DashboardSearch,
  DashboardSearchGroupByMode,
} from "../../types/dashboard";
import { groupBy } from "lodash";
import { useEffect, useState } from "react";

const defaultSearch: DashboardSearch = {
  value: "",
  groupBy: {
    value: "tag",
    tag: "service",
  },
};

const buildDashboardSearch = (
  searchParams: URLSearchParams,
  defaults?: DashboardSearch
): DashboardSearch => {
  const search: DashboardSearch = {
    value: "",
    groupBy: {
      value: "tag",
      tag: null,
    },
  };

  if (searchParams.has("search")) {
    search.value = searchParams.get("search") as string;
  } else if (defaults && defaults.value) {
    search.value = defaults.value;
  } else {
    search.value = defaultSearch.value;
  }

  if (searchParams.has("group_by")) {
    search.groupBy.value = searchParams.get(
      "group_by"
    ) as DashboardSearchGroupByMode;
  } else if (defaults && defaults.groupBy && defaults.groupBy.value) {
    search.groupBy.value = defaults.groupBy.value;
  } else {
    search.groupBy.value = defaultSearch.groupBy.value;
  }

  if (searchParams.has("tag")) {
    search.groupBy.tag = searchParams.get("tag");
  } else if (defaults && defaults.groupBy && defaults.groupBy.tag) {
    search.groupBy.tag = defaults.groupBy.tag;
  } else {
    search.groupBy.tag = defaultSearch.groupBy.tag;
  }

  if (search.groupBy.value === "mod") {
    search.groupBy.tag = null;
  }

  return search;
};

const useDashboardSearch = (
  searchParams: URLSearchParams,
  defaults?: DashboardSearch
) => {
  const [search, setSearch] = useState(
    buildDashboardSearch(searchParams, defaults)
  );

  useEffect(() => {
    setSearch(buildDashboardSearch(searchParams, defaults));
  }, [searchParams, defaults]);

  return { search, groupBy };
};

export default useDashboardSearch;
