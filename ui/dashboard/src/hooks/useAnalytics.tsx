import usePrevious from "./usePrevious";
import {
  AvailableDashboard,
  CloudDashboardIdentityMetadata,
  CloudDashboardWorkspaceMetadata,
  ModDashboardMetadata,
  useDashboard,
} from "./useDashboard";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
} from "react";
import { get } from "lodash";
import { useTheme } from "./useTheme";

interface AnalyticsProperties {
  [key: string]: any;
}

interface IAnalyticsContext {
  reset: () => void;
  track: (string, AnalyticsProperties) => void;
}

interface SelectedDashboardStates {
  selectedDashboard: AvailableDashboard | null;
}

const AnalyticsContext = createContext<IAnalyticsContext>({
  reset: () => {},
  track: () => {},
});

const useAnalyticsProvider = () => {
  // const location = useLocation();
  const { metadata, metadataLoaded, selectedDashboard } = useDashboard();
  const { localStorageTheme, theme } = useTheme();
  const [identity, setIdentity] =
    useState<CloudDashboardIdentityMetadata | null>(null);
  const [workspace, setWorkspace] =
    useState<CloudDashboardWorkspaceMetadata | null>(null);

  const reset = useCallback(() => {
    // @ts-ignore
    window.heap && window.heap.resetIdentity();
  }, []);

  const track = useCallback(
    (event, properties) => {
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
      // console.log("Tracking", { event, properties: finalProperties });
      // @ts-ignore
      window.heap && window.heap.track(event, properties);
    },
    [identity, workspace, localStorageTheme, theme]
  );

  useEffect(() => {
    if (!metadataLoaded) {
      return;
    }

    const cloudMetadata = metadata.cloud;

    if (!cloudMetadata) {
      reset();
    }

    const actor = cloudMetadata?.actor;
    const identity = cloudMetadata?.identity;
    const workspace = cloudMetadata?.workspace;

    setIdentity(identity ? identity : null);
    setWorkspace(workspace ? workspace : null);

    if (actor) {
      // @ts-ignore
      window.heap && window.heap.identify(actor.id);
    } else {
      reset();
    }
  }, [metadataLoaded, metadata]);

  // @ts-ignore
  const previousSelectedDashboardStates: SelectedDashboardStates = usePrevious({
    selectedDashboard,
  });

  useEffect(() => {
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
        mod = get(metadata, "mod", {} as ModDashboardMetadata);
      } else {
        mod = get(
          metadata,
          `installed_mods["${selectedDashboard.mod_full_name}"]`,
          {} as ModDashboardMetadata
        );
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
  }, [metadata, previousSelectedDashboardStates, selectedDashboard]);

  return {
    reset,
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
