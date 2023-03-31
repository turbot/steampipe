import { DashboardRunState } from "../../../../types";
import { getColumn } from "../../../../utils/data";
import { LeafNodeData } from "../../common";
import { SelectInputOption, SelectOption } from "../types";
import { useMemo } from "react";

const useSelectInputValues = (
  options: SelectInputOption[] | undefined,
  data: LeafNodeData | undefined,
  status: DashboardRunState
) => {
  // Get the options for the select
  return useMemo<SelectOption[]>(() => {
    // If no options defined at all
    if (
      ((!options || options.length === 0) &&
        (!data || !data.columns || !data.rows)) ||
      // This property is only present in workspaces >=v0.16.x
      (status !== undefined && status !== "complete")
    ) {
      return [];
    }

    if (data) {
      const labelCol = getColumn(data.columns, "label");
      const valueCol = getColumn(data.columns, "value");
      const tagsCol = getColumn(data.columns, "tags");

      if (!labelCol || !valueCol) {
        return [];
      }

      return data.rows.map((row) => {
        const label = row[labelCol.name];
        const value = row[valueCol.name];
        return {
          label: !!label ? label.toString() : "",
          value: !!value ? value.toString() : null,
          tags: tagsCol ? row[tagsCol.name] : {},
        };
      });
    } else if (options) {
      return options.map((option) => ({
        label: option.label || option.name,
        value: option.name,
        tags: {},
      }));
    } else {
      return [];
    }
  }, [options, data, status]);
};

export default useSelectInputValues;
