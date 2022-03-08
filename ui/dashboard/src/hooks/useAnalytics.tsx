import {
  AvailableDashboard,
  CloudDashboardActorMetadata,
  CloudDashboardIdentityMetadata,
  CloudDashboardWorkspaceMetadata,
  useDashboard,
} from "./useDashboard";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
} from "react";
import usePrevious from "./usePrevious";

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
  const [actor, setActor] = useState<CloudDashboardActorMetadata | null>(null);
  const [identity, setIdentity] =
    useState<CloudDashboardIdentityMetadata | null>(null);
  const [workspace, setWorkspace] =
    useState<CloudDashboardWorkspaceMetadata | null>(null);

  const reset = useCallback(() => {
    // @ts-ignore
    window.heap && window.heap.resetIdentity();
    // }, [analytics]);
  }, []);

  const track = useCallback((event, properties) => {
    // @ts-ignore
    window.heap && window.heap.track(event, properties);
  }, []);

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

    setActor(actor ? actor : null);
    setIdentity(identity ? identity : null);
    setWorkspace(workspace ? workspace : null);

    if (actor) {
      // @ts-ignore
      window.heap && window.heap.identify(actor.id);
    } else {
      reset();
    }
  }, [metadataLoaded, metadata]);

  const previousSelectedDashboardStates: SelectedDashboardStates | undefined =
    usePrevious({
      selectedDashboard,
    });

  useEffect(() => {
    console.log({
      actor,
      identity,
      workspace,
      previousSelectedDashboardStates,
      selectedDashboard,
    });
  }, [
    actor,
    identity,
    workspace,
    previousSelectedDashboardStates,
    selectedDashboard,
  ]);

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
