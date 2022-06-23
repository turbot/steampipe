import { DashboardDataMode } from "../../types/dashboard";
import { useEffect, useState } from "react";

const buildDashboardSnapshotInfoFromSearchParams = (
  searchParams: URLSearchParams,
  featureFlags: string[]
): string | null => {
  if (!featureFlags.includes("snapshots")) {
    return null;
  }

  if (searchParams.has("snapshot_id")) {
    return searchParams.get("snapshot_id");
  }

  return null;
};

const useDashboardSnapshot = (
  searchParams: URLSearchParams,
  dashboardState: any,
  dataMode: DashboardDataMode,
  featureFlags: string[]
) => {
  const [snapshotId, setSnapshotId] = useState<string | null>(
    buildDashboardSnapshotInfoFromSearchParams(searchParams, featureFlags)
  );

  useEffect(() => {
    setSnapshotId(
      buildDashboardSnapshotInfoFromSearchParams(searchParams, featureFlags)
    );
  }, [featureFlags, searchParams, setSnapshotId]);

  return { id: snapshotId };
};

export default useDashboardSnapshot;
