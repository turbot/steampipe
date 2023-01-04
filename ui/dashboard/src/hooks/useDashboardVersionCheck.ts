import { IDashboardContext } from "../types";
import { useEffect } from "react";

const useDashboardVersionCheck = (state: IDashboardContext) => {
  console.log(state, process.env.REACT_APP_VERSION);
  useEffect(() => {
    let cliVersion: string | null = "";
    let uiVersion: string | null = "";
    let mismatchedVersions = false;
    if (state.versionMismatchCheck) {
      const cliVersionRaw = state.metadata?.cli?.version;
      const uiVersionRaw = process.env.REACT_APP_VERSION;
      const hasVersionsSet = !!cliVersionRaw && !!uiVersionRaw;
      cliVersion = !!cliVersionRaw
        ? cliVersionRaw.startsWith("v")
          ? cliVersionRaw.substring(1)
          : cliVersionRaw
        : null;
      uiVersion = !!uiVersionRaw
        ? uiVersionRaw.startsWith("v")
          ? uiVersionRaw.substring(1)
          : uiVersionRaw
        : null;
      mismatchedVersions = hasVersionsSet && cliVersion !== uiVersion;

      const searchParams = new URLSearchParams(window.location.search);

      if (mismatchedVersions && cliVersionRaw) {
        searchParams.set("version", cliVersionRaw);
        console.log(searchParams.toString());
        const url = new URL(window.location.href);
        console.log(url);
        console.log({ cliVersion, uiVersion, mismatchedVersions });
        window.location.replace(`${window.location.origin}?${searchParams}`);
      }
    }
  }, [state]);
};

export default useDashboardVersionCheck;
