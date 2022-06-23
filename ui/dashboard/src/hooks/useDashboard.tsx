import { createContext, useCallback, useReducer } from "react";
import { IActions } from "../types/dashboard";
import { useSearchParams } from "react-router-dom";

// State
// Socket
// Query Params
// History
// Inputs

export type DashboardDataMode = "live" | "snapshot";

interface IDashboardContext {
  dispatch(action: DashboardAction): void;

  dataMode: DashboardDataMode;
  snapshotId: string | null;

  refetchDashboard: boolean;
  selectedSnapshot: DashboardSnapshot | null;
  lastChangedInput: string | null;
}

const DashboardActions: IActions = {
  CLEAR_DASHBOARD_INPUTS: "clear_dashboard_inputs",
  CLEAR_SNAPSHOT: "clear_snapshot",
  CONTROL_COMPLETE: "control_complete",
  CONTROL_ERROR: "control_error",
  DELETE_DASHBOARD_INPUT: "delete_dashboard_input",
  EXECUTION_COMPLETE: "execution_complete",
  EXECUTION_ERROR: "execution_error",
  EXECUTION_STARTED: "execution_started",
  INPUT_VALUES_CLEARED: "input_values_cleared",
  LEAF_NODE_COMPLETE: "leaf_node_complete",
  LEAF_NODE_PROGRESS: "leaf_node_progress",
  SELECT_DASHBOARD: "select_dashboard",
  SELECT_PANEL: "select_panel",
  SELECT_SNAPSHOT: "select_snapshot",
  SET_DASHBOARD: "set_dashboard",
  SET_DASHBOARD_INPUT: "set_dashboard_input",
  SET_DASHBOARD_INPUTS: "set_dashboard_inputs",
  SET_DASHBOARD_SEARCH_VALUE: "set_dashboard_search_value",
  SET_DASHBOARD_SEARCH_GROUP_BY: "set_dashboard_search_group_by",
  SET_DATA_MODE: "set_data_mode",
  SET_REFETCH_DASHBOARD: "set_refetch_dashboard",
  SET_SNAPSHOT: "set_snapshot",
  WORKSPACE_ERROR: "workspace_error",
};

const dashboardActions = Object.values(DashboardActions);

// https://github.com/microsoft/TypeScript/issues/28046
export type ElementType<T extends ReadonlyArray<unknown>> =
  T extends ReadonlyArray<infer ElementType> ? ElementType : never;

export type DashboardActionType = ElementType<typeof dashboardActions>;

export interface DashboardAction {
  type: DashboardActionType;
  [key: string]: any;
}

interface DashboardInputs {
  [name: string]: string;
}

interface DashboardVariables {
  [name: string]: any;
}

export interface DashboardSnapshot {
  id: string;
  dashboard_name: string;
  start_time: string;
  end_time: string;
  lineage: string;
  schema_version: string;
  search_path: string;
  variables: DashboardVariables;
  inputs: DashboardInputs;
}

interface DashboardProviderProps {
  children: null | JSX.Element | JSX.Element[];
  eventHooks?: {};
  featureFlags?: string[];
  socketFactory?: () => WebSocket;
  stateDefaults?: {};
}

function reducer(state, action) {
  switch (action.type) {
    case DashboardActions.CLEAR_SNAPSHOT:
      return {
        ...state,
        selectedSnapshot: null,
        snapshotId: null,
        dataMode: "live",
      };
    case DashboardActions.SELECT_SNAPSHOT:
      return {
        ...state,
        selectedSnapshot: action.snapshot,
        snapshotId: action.snapshot.id,
        dataMode: "snapshot",
      };
    case DashboardActions.SET_DATA_MODE:
      return {
        ...state,
        dataMode: action.dataMode,
      };
    case DashboardActions.SET_REFETCH_DASHBOARD:
      return {
        ...state,
        refetchDashboard: true,
      };
    case DashboardActions.SET_DASHBOARD:
      return {
        ...state,
        dashboard: action.dashboard,
      };
    case DashboardActions.CLEAR_DASHBOARD_INPUTS:
      return {
        ...state,
        selectedDashboardInputs: {},
        lastChangedInput: null,
        recordInputsHistory: !!action.recordInputsHistory,
      };
    case DashboardActions.DELETE_DASHBOARD_INPUT:
      const { [action.name]: toDelete, ...rest } =
        state.selectedDashboardInputs;
      return {
        ...state,
        selectedDashboardInputs: {
          ...rest,
        },
        lastChangedInput: action.name,
        recordInputsHistory: !!action.recordInputsHistory,
      };
    case DashboardActions.SET_DASHBOARD_INPUT:
      return {
        ...state,
        selectedDashboardInputs: {
          ...state.selectedDashboardInputs,
          [action.name]: action.value,
        },
        lastChangedInput: action.name,
        recordInputsHistory: !!action.recordInputsHistory,
      };
    case DashboardActions.SET_DASHBOARD_INPUTS:
      return {
        ...state,
        selectedDashboardInputs: action.value,
        lastChangedInput: null,
        recordInputsHistory: !!action.recordInputsHistory,
      };
    case DashboardActions.INPUT_VALUES_CLEARED: {
      // We're not expecting execution events for this ID
      if (action.execution_id !== state.execution_id) {
        return state;
      }
      const newSelectedDashboardInputs = { ...state.selectedDashboardInputs };
      for (const input of action.cleared_inputs || []) {
        delete newSelectedDashboardInputs[input];
      }
      return {
        ...state,
        selectedDashboardInputs: newSelectedDashboardInputs,
        lastChangedInput: null,
        recordInputsHistory: false,
      };
    }
    default:
      console.warn(`Unsupported action ${action.type}`, action);
      return state;
  }
}

const buildSelectedDashboardInputsFromSearchParams = (searchParams) => {
  const selectedDashboardInputs = {};
  // @ts-ignore
  for (const entry of searchParams.entries()) {
    if (!entry[0].startsWith("input")) {
      continue;
    }
    selectedDashboardInputs[entry[0]] = entry[1];
  }
  return selectedDashboardInputs;
};

const getInitialState = (searchParams) => {
  return {
    dataMode: searchParams.get("mode") || "live",
    snapshotId: searchParams.has("snapshot_id")
      ? searchParams.get("snapshot_id")
      : null,
    refetchDashboard: false,
    error: null,

    panelsMap: {},
    dashboard: null,
    selectedPanel: null,
    selectedDashboardInputs:
      buildSelectedDashboardInputsFromSearchParams(searchParams),
    selectedSnapshot: null,
    lastChangedInput: null,

    sqlDataMap: {},

    execution_id: null,

    progress: 0,
  };
};

const DashboardContext = createContext<IDashboardContext | null>(null);

const DashboardProvider = ({ children }: DashboardProviderProps) => {
  const [searchParams] = useSearchParams();
  const [state, dispatchInner] = useReducer(
    reducer,
    getInitialState(searchParams)
  );
  const dispatch = useCallback((action) => {
    // console.log(action.type, action);
    dispatchInner(action);
  }, []);
  // const { dashboard_name } = useParams();
  // const { ready: socketReady, send: sendSocketMessage } = useDashboardWebSocket(
  //   dispatch,
  //   socketFactory,
  //   eventHooks
  // );

  // const navigationType = useNavigationType();

  // Keep track of the previous selected dashboard and inputs
  // const previousSelectedDashboardStates: SelectedDashboardStates | undefined =
  //   usePrevious({
  //     searchParams,
  //     dashboard_name,
  //     dataMode: state.dataMode,
  //     refetchDashboard: state.refetchDashboard,
  //     search: state.search,
  //     selectedDashboard: state.selectedDashboard,
  //     selectedDashboardInputs: state.selectedDashboardInputs,
  //     selectedSnapshot: state.selectedSnapshot,
  //   });

  // Initial sync into URL
  // useEffect(() => {
  //   if (
  //     !featureFlags.includes("snapshots") ||
  //     (searchParams.has("mode") && searchParams.get("mode") === state.dataMode)
  //   ) {
  //     return;
  //   }
  //   searchParams.set("mode", state.dataMode);
  //   setSearchParams(searchParams, { replace: true });
  // }, [featureFlags, searchParams, setSearchParams, state.dataMode]);

  // useEffect(() => {
  //   if (featureFlags.includes("snapshots") && state.selectedSnapshot) {
  //     searchParams.set("snapshot_id", state.selectedSnapshot.id);
  //     setSearchParams(searchParams, { replace: true });
  //   }
  // }, [featureFlags, searchParams, setSearchParams, state.selectedSnapshot]);

  // useEffect(() => {
  //   if (
  //     featureFlags.includes("snapshots") &&
  //     state.dataMode === "live" &&
  //     searchParams.has("snapshot_id")
  //   ) {
  //     searchParams.delete("snapshot_id");
  //     setSearchParams(searchParams, { replace: true });
  //   }
  // }, [featureFlags, searchParams, setSearchParams, state.dataMode]);

  // Ensure that on history pop / push we sync the new values into state
  // useEffect(() => {
  //   if (navigationType !== "POP" && navigationType !== "PUSH") {
  //     return;
  //   }
  //   if (location.key === "default") {
  //     return;
  //   }
  //
  //   // If we've just popped or pushed from one dashboard to another, then we don't want to add the search to the URL
  //   // as that will show the dashboard list, but we want to see the dashboard that we came from / went to previously.
  //   const goneFromDashboardToDashboard =
  //     // @ts-ignore
  //     previousSelectedDashboardStates?.dashboard_name &&
  //     dashboard_name &&
  //     // @ts-ignore
  //     previousSelectedDashboardStates.dashboard_name !== dashboard_name;
  //
  //   const search = searchParams.get("search") || "";
  //   const groupBy =
  //     searchParams.get("group_by") ||
  //     get(stateDefaults, "search.groupBy.value", "tag");
  //   const tag =
  //     searchParams.get("tag") ||
  //     get(stateDefaults, "search.groupBy.tag", "service");
  //   const dataMode = searchParams.has("mode")
  //     ? searchParams.get("mode")
  //     : "live";
  //   const inputs = buildSelectedDashboardInputsFromSearchParams(searchParams);
  //   dispatch({
  //     type: DashboardActions.SET_DASHBOARD_SEARCH_VALUE,
  //     value: goneFromDashboardToDashboard ? "" : search,
  //   });
  //   dispatch({
  //     type: DashboardActions.SET_DASHBOARD_SEARCH_GROUP_BY,
  //     value: groupBy,
  //     tag,
  //   });
  //   dispatch({
  //     type: DashboardActions.SET_DASHBOARD_INPUTS,
  //     value: inputs,
  //     recordInputsHistory: false,
  //   });
  //   if (featureFlags.includes("snapshots")) {
  //     dispatch({
  //       type: DashboardActions.SET_DATA_MODE,
  //       dataMode,
  //     });
  //   }
  // }, [
  //   dashboard_name,
  //   dispatch,
  //   featureFlags,
  //   location,
  //   navigationType,
  //   previousSelectedDashboardStates,
  //   searchParams,
  //   stateDefaults,
  // ]);

  // useEffect(() => {
  //   // If no search params have changed
  //   if (
  //     previousSelectedDashboardStates &&
  //     // @ts-ignore
  //     previousSelectedDashboardStates?.dashboard_name === dashboard_name &&
  //     // @ts-ignore
  //     previousSelectedDashboardStates.dataMode === state.dataMode &&
  //     // @ts-ignore
  //     previousSelectedDashboardStates.search.value === state.search.value &&
  //     // @ts-ignore
  //     previousSelectedDashboardStates.search.groupBy.value ===
  //       state.search.groupBy.value &&
  //     // @ts-ignore
  //     previousSelectedDashboardStates.search.groupBy.tag ===
  //       state.search.groupBy.tag &&
  //     // @ts-ignore
  //     previousSelectedDashboardStates.searchParams.toString() ===
  //       searchParams.toString()
  //   ) {
  //     return;
  //   }
  //
  //   const {
  //     value: searchValue,
  //     groupBy: { value: groupByValue, tag },
  //   } = state.search;
  //
  //   if (dashboard_name) {
  //     // Only set group_by and tag if we have a search
  //     if (searchValue) {
  //       searchParams.set("search", searchValue);
  //       searchParams.set("group_by", groupByValue);
  //
  //       if (groupByValue === "mod") {
  //         searchParams.delete("tag");
  //       } else if (groupByValue === "tag") {
  //         searchParams.set("tag", tag);
  //       } else {
  //         searchParams.delete("group_by");
  //         searchParams.delete("tag");
  //       }
  //     } else {
  //       searchParams.delete("search");
  //       searchParams.delete("group_by");
  //       searchParams.delete("tag");
  //     }
  //   } else {
  //     if (searchValue) {
  //       searchParams.set("search", searchValue);
  //     } else {
  //       searchParams.delete("search");
  //     }
  //
  //     searchParams.set("group_by", groupByValue);
  //
  //     if (groupByValue === "mod") {
  //       searchParams.delete("tag");
  //     } else if (groupByValue === "tag") {
  //       searchParams.set("tag", tag);
  //     } else {
  //       searchParams.delete("group_by");
  //       searchParams.delete("tag");
  //     }
  //   }
  //
  //   if (featureFlags.includes("snapshots")) {
  //     searchParams.set("mode", state.dataMode);
  //   }
  //   setSearchParams(searchParams, { replace: true });
  // }, [
  //   dashboard_name,
  //   featureFlags,
  //   previousSelectedDashboardStates,
  //   searchParams,
  //   setSearchParams,
  //   state.dataMode,
  //   state.search,
  // ]);

  // useEffect(() => {
  //   // This effect will send events over websockets and depends on there being a dashboard selected
  //   if (!socketReady || !state.selectedDashboard) {
  //     return;
  //   }
  //
  //   // If we didn't previously have a dashboard selected in state (e.g. you've gone from home page
  //   // to a report, or it's first load), or the selected dashboard has been changed, select that
  //   // report over the socket
  //   if (
  //     state.dataMode === "live" &&
  //     (!previousSelectedDashboardStates ||
  //       // @ts-ignore
  //       !previousSelectedDashboardStates.selectedDashboard ||
  //       state.selectedDashboard.full_name !==
  //         // @ts-ignore
  //         previousSelectedDashboardStates.selectedDashboard.full_name ||
  //       // @ts-ignore
  //       (!previousSelectedDashboardStates.refetchDashboard &&
  //         state.refetchDashboard))
  //   ) {
  //     sendSocketMessage({
  //       action: SocketActions.CLEAR_DASHBOARD,
  //     });
  //     sendSocketMessage({
  //       action: SocketActions.SELECT_DASHBOARD,
  //       payload: {
  //         dashboard: {
  //           full_name: state.selectedDashboard.full_name,
  //         },
  //         input_values: state.selectedDashboardInputs,
  //       },
  //     });
  //     return;
  //   }
  //   // Else if we did previously have a dashboard selected in state and the
  //   // inputs have changed, then update the inputs over the socket
  //   if (
  //     state.dataMode === "live" &&
  //     previousSelectedDashboardStates &&
  //     // @ts-ignore
  //     previousSelectedDashboardStates.selectedDashboard &&
  //     !isEqual(
  //       // @ts-ignore
  //       previousSelectedDashboardStates.selectedDashboardInputs,
  //       state.selectedDashboardInputs
  //     )
  //   ) {
  //     sendSocketMessage({
  //       action: SocketActions.INPUT_CHANGED,
  //       payload: {
  //         dashboard: {
  //           full_name: state.selectedDashboard.full_name,
  //         },
  //         changed_input: state.lastChangedInput,
  //         input_values: state.selectedDashboardInputs,
  //       },
  //     });
  //   }
  // }, [
  //   previousSelectedDashboardStates,
  //   sendSocketMessage,
  //   socketReady,
  //   state.selectedDashboard,
  //   state.selectedDashboardInputs,
  //   state.lastChangedInput,
  //   state.dataMode,
  //   state.refetchDashboard,
  // ]);

  // useEffect(() => {
  //   // This effect will send events over websockets and depends on there being no dashboard selected
  //   if (!socketReady || state.selectedDashboard) {
  //     return;
  //   }
  //
  //   // If we've gone from having a report selected, to having nothing selected, clear the dashboard state
  //   if (
  //     previousSelectedDashboardStates &&
  //     // @ts-ignore
  //     previousSelectedDashboardStates.selectedDashboard
  //   ) {
  //     sendSocketMessage({
  //       action: SocketActions.CLEAR_DASHBOARD,
  //     });
  //   }
  // }, [
  //   previousSelectedDashboardStates,
  //   sendSocketMessage,
  //   socketReady,
  //   state.selectedDashboard,
  // ]);

  // useEffect(() => {
  //   // Don't do anything as this is handled elsewhere
  //   if (navigationType === "POP" || navigationType === "PUSH") {
  //     return;
  //   }
  //
  //   if (!previousSelectedDashboardStates) {
  //     return;
  //   }
  //
  //   if (
  //     isEqual(
  //       state.selectedDashboardInputs,
  //       // @ts-ignore
  //       previousSelectedDashboardStates.selectedDashboardInputs
  //     )
  //   ) {
  //     return;
  //   }
  //
  //   // Only record history when it's the same report before and after and the inputs have changed
  //   const shouldRecordHistory =
  //     state.recordInputsHistory &&
  //     // @ts-ignore
  //     !!previousSelectedDashboardStates.selectedDashboard &&
  //     !!state.selectedDashboard &&
  //     // @ts-ignore
  //     previousSelectedDashboardStates.selectedDashboard.full_name ===
  //       state.selectedDashboard.full_name;
  //
  //   // Sync params into the URL
  //   const newParams = {
  //     ...state.selectedDashboardInputs,
  //   };
  //   if (featureFlags.includes("snapshots")) {
  //     newParams.mode = state.dataMode;
  //   }
  //   setSearchParams(newParams, {
  //     replace: !shouldRecordHistory,
  //   });
  // }, [
  //   featureFlags,
  //   navigationType,
  //   previousSelectedDashboardStates,
  //   setSearchParams,
  //   state.dataMode,
  //   state.recordInputsHistory,
  //   state.selectedDashboard,
  //   state.selectedDashboardInputs,
  // ]);

  return (
    <DashboardContext.Provider
      value={{
        ...state,
        dispatch,
      }}
    >
      {children}
    </DashboardContext.Provider>
  );
};

export { DashboardContext, DashboardProvider };
