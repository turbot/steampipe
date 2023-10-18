import get from "lodash/get";
import isNumber from "lodash/isNumber";
import { CardProperties } from "../Card";
import {
  DashboardPanelType,
  DashboardRunState,
  PanelDefinition,
} from "../../../types";
import { getColumn, hasData } from "../../../utils/data";
import { getIconForType } from "../../../utils/card";
import { IDiffProperties, IPanelDiff } from "./types";
import { isNumericCol, LeafNodeData } from "../common";

export interface CardDiffState extends IPanelDiff {
  value?: number;
  value_percent?: "infinity" | number;
  direction: "none" | "up" | "down";
  status?: "ok" | "alert" | "severity" | null;
}

export interface CardState {
  loading: boolean;
  label: string | null;
  value: any | null;
  value_number: number | null;
  type: CardType;
  icon: string | null;
  href: string | null;
  diff?: CardDiffState;
}

export type CardDataFormat = "simple" | "formal";

export type CardType = "alert" | "info" | "ok" | "severity" | "table" | null;

export class CardDataProcessor {
  constructor() {}

  getDefaultState = (
    status: DashboardRunState,
    properties: CardProperties,
    display_type: CardType | undefined,
  ): CardState => {
    return {
      loading: status === "running",
      label: properties.label || null,
      value: isNumber(properties.value)
        ? properties.value.toLocaleString()
        : properties.value || null,
      value_number: isNumber(properties.value) ? properties.value : null,
      type: display_type || null,
      icon: getIconForType(display_type, properties.icon),
      href: properties.href || null,
    };
  };

  buildCardState(
    data: LeafNodeData | undefined,
    diff_panel: PanelDefinition | undefined,
    display_type: CardType | undefined,
    properties: CardProperties,
    status: DashboardRunState,
  ): CardState {
    if (!data || !hasData(data)) {
      const state = this.getDefaultState(status, properties, display_type);
      if (!!diff_panel && !!diff_panel.properties) {
        const diffState = this.getDefaultState(
          status,
          diff_panel.properties,
          display_type,
        );
        state.diff = this.diff(properties, state, diffState) as CardDiffState;
      }
      return state;
    }

    const state = this.parseData(data, display_type, properties);

    if (!!diff_panel && !!diff_panel.data) {
      const diffState = this.parseData(
        diff_panel.data,
        display_type,
        properties,
      );
      state.diff = this.diff(properties, state, diffState) as CardDiffState;
    }

    return state;
  }

  parseData(
    data: LeafNodeData,
    display_type: CardType | undefined,
    properties: CardProperties,
  ): CardState {
    const dataFormat = this.getDataFormat(data);
    if (dataFormat === "simple") {
      const firstCol = data.columns[0];
      const isNumericValue = isNumericCol(firstCol.data_type);
      const row = data.rows[0];
      const value = row[firstCol.name];
      return {
        loading: false,
        label: firstCol.name,
        value:
          value !== null && value !== undefined && isNumericValue
            ? value.toLocaleString()
            : value,
        value_number: isNumericValue && isNumber(value) ? value : null,
        type: display_type || null,
        icon: getIconForType(display_type, properties.icon),
        href: properties.href || null,
      };
    } else {
      const formalLabel = get(data, "rows[0].label", null);
      const formalValue = get(data, `rows[0].value`, null);
      const formalType = get(data, `rows[0].type`, null);
      const formalIcon = get(data, `rows[0].icon`, null);
      const formalHref = get(data, `rows[0].href`, null);
      const valueCol = getColumn(data.columns, "value");
      const isNumericValue = !!valueCol && isNumericCol(valueCol.data_type);
      return {
        loading: false,
        label: formalLabel,
        value:
          formalValue !== null && formalValue !== undefined && isNumericValue
            ? formalValue.toLocaleString()
            : formalValue,
        value_number:
          formalValue && isNumericValue && isNumber(formalValue)
            ? formalValue
            : null,
        type: formalType || display_type || null,
        icon: getIconForType(
          formalType || display_type,
          formalIcon || properties.icon,
        ),
        href: formalHref || properties.href || null,
      };
    }
  }

  diff(
    properties: IDiffProperties,
    state: CardState,
    previous_state: CardState,
  ): CardDiffState {
    // If the columns aren't numeric then we can't diff...
    if (state.value_number === null || previous_state.value_number === null) {
      return {
        direction: "none",
      };
    }

    const direction =
      state.value_number > previous_state.value_number
        ? "up"
        : state.value_number === previous_state.value_number
        ? "none"
        : "down";

    let value: number;
    let value_percent: "infinity" | number;
    let status: "ok" | "alert" | "severity" | null = null;
    if (direction === "up") {
      value = state.value_number - previous_state.value_number;
      status =
        state.type === "alert"
          ? "alert"
          : state.type === "ok"
          ? "ok"
          : state.type === "severity"
          ? "severity"
          : null;
    } else if (direction === "down") {
      value = previous_state.value_number - state.value_number;
      status =
        state.type === "alert"
          ? "ok"
          : state.type === "ok"
          ? "alert"
          : state.type === "severity"
          ? "ok"
          : null;
    } else {
      value = 0;
    }

    if (state.value_number === 0) {
      value_percent = "infinity";
    } else if (value === 0) {
      value_percent = 0;
    } else {
      value_percent = Math.ceil((value / state.value_number) * 100);
    }

    return {
      value,
      value_percent,
      direction,
      status,
    };
  }

  getDataFormat = (data: LeafNodeData | undefined): CardDataFormat => {
    if (!!data && data.columns.length > 1) {
      return "formal";
    }
    return "simple";
  };

  get panel_type(): DashboardPanelType {
    return "card";
  }
}
