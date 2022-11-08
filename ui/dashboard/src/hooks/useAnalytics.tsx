import get from "lodash/get";
import usePrevious from "./usePrevious";
import {
  AvailableDashboard,
  CloudDashboardIdentityMetadata,
  CloudDashboardWorkspaceMetadata,
  DashboardMetadata,
  ModDashboardMetadata,
} from "../types";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
} from "react";
import { useTheme } from "./useTheme";

interface AnalyticsProperties {
  [key: string]: any;
}

interface IAnalyticsContext {
  reset: () => void;
  track: (string, properties: AnalyticsProperties) => void;
}

interface SelectedDashboardStates {
  selectedDashboard: AvailableDashboard | null;
}

const AnalyticsContext = createContext<IAnalyticsContext>({
  reset: () => {},
  track: () => {},
});

const useAnalyticsProvider = () => {
  const { localStorageTheme, theme } = useTheme();
  const [enabled, setEnabled] = useState<boolean>(true);
  const [identity, setIdentity] =
    useState<CloudDashboardIdentityMetadata | null>(null);
  const [workspace, setWorkspace] =
    useState<CloudDashboardWorkspaceMetadata | null>(null);
  const [metadata, setMetadata] = useState<DashboardMetadata | null>(null);
  const [selectedDashboard, setSelectedDashboard] =
    useState<AvailableDashboard | null>(null);
  const [initialised, setInitialised] = useState(false);

  const identify = useCallback((actor) => {
    // @ts-ignore
    if (window.heap && window.heap.identify) {
      // @ts-ignore
      window.heap.identify(actor.id);
    }
  }, []);

  const reset = useCallback(() => {
    // @ts-ignore
    if (window.heap && window.heap.resetIdentity) {
      // @ts-ignore
      window.heap.resetIdentity();
    }
  }, []);

  const track = useCallback(
    (event, properties) => {
      if (!initialised || !enabled) {
        return;
      }
      const additionalProperties = {
        theme: theme.name,
        using_system_theme: !localStorageTheme,
      };
      if (identity) {
        additionalProperties["identity.type"] = identity.type;
        additionalProperties["identity.id"] = identity.id;
        additionalProperties["identity.handle"] = identity.handle;
      }
      if (workspace) {
        additionalProperties["workspace.id"] = workspace.id;
        additionalProperties["workspace.handle"] = workspace.handle;
      }
      const finalProperties = {
        ...additionalProperties,
        ...properties,
      };
      // @ts-ignore
      if (window.heap && window.heap.track) {
        // @ts-ignore
        window.heap.track(event, finalProperties);
      }
    },
    [enabled, initialised, identity, workspace, localStorageTheme, theme.name]
  );

  useEffect(() => {
    if (initialised || !metadata) {
      return;
    }

    const enabled =
      process.env.NODE_ENV === "production" &&
      metadata.telemetry === "info" &&
      !!process.env.REACT_APP_HEAP_ID;

    // @ts-ignore
    if (enabled && window.heap && window.heap.load) {
      // @ts-ignore
      window.heap.load(process.env.REACT_APP_HEAP_ID);
    }

    setEnabled(enabled);
    setInitialised(true);
  }, [initialised, metadata, setEnabled, setInitialised]);

  useEffect(() => {
    if (!metadata || !initialised) {
      return;
    }

    const cloudMetadata = metadata.cloud;

    const identity = cloudMetadata?.identity;
    const workspace = cloudMetadata?.workspace;

    setIdentity(identity ? identity : null);
    setWorkspace(workspace ? workspace : null);

    const actor = cloudMetadata?.actor;

    if (actor && enabled) {
      identify(actor);
    } else if (enabled) {
      reset();
    }
  }, [
    enabled,
    identify,
    initialised,
    metadata,
    reset,
    setIdentity,
    setWorkspace,
  ]);

  // @ts-ignore
  const previousSelectedDashboardStates: SelectedDashboardStates = usePrevious({
    selectedDashboard,
  });

  useEffect(() => {
    if (!enabled || !metadata) {
      return;
    }

    if (
      ((!previousSelectedDashboardStates ||
        !previousSelectedDashboardStates.selectedDashboard) &&
        selectedDashboard) ||
      (previousSelectedDashboardStates &&
        previousSelectedDashboardStates.selectedDashboard &&
        selectedDashboard &&
        previousSelectedDashboardStates.selectedDashboard.full_name !==
          selectedDashboard?.full_name)
    ) {
      let mod: ModDashboardMetadata;
      if (selectedDashboard.mod_full_name === metadata.mod.full_name) {
        mod = get(metadata, "mod", {}) as ModDashboardMetadata;
      } else {
        mod = get(
          metadata,
          `installed_mods["${selectedDashboard.mod_full_name}"]`,
          {}
        ) as ModDashboardMetadata;
      }
      track("cli.ui.dashboard.select", {
        "mod.title": mod
          ? mod.title
            ? mod.title
            : mod.short_name
          : selectedDashboard.mod_full_name,
        "mod.name": mod ? mod.short_name : selectedDashboard.mod_full_name,
        dashboard: selectedDashboard.short_name,
      });
    }
  }, [
    enabled,
    metadata,
    previousSelectedDashboardStates,
    selectedDashboard,
    track,
  ]);

  return {
    reset,
    setMetadata,
    setSelectedDashboard,
    track,
  };
};

const AnalyticsProvider = ({ children }) => {
  const analytics = useAnalyticsProvider();

  return (
    <AnalyticsContext.Provider value={analytics as IAnalyticsContext}>
      {children}
    </AnalyticsContext.Provider>
  );
};

const useAnalytics = () => {
  const context = useContext(AnalyticsContext);
  if (context === undefined) {
    throw new Error("useAnalytics must be used within an AnalyticsContext");
  }
  return context;
};

export default useAnalytics;

export { AnalyticsProvider };
