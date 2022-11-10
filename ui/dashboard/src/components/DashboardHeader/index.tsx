import DashboardSearch from "../DashboardSearch";
import DashboardTagGroupSelect from "../DashboardTagGroupSelect";
import OpenSnapshotButton from "../OpenSnapshotButton";
import SaveSnapshotButton from "../SaveSnapshotButton";
import SteampipeLogo from "./SteampipeLogo";
import ThemeToggle from "../ThemeToggle";
import { classNames } from "../../utils/styles";
import { useDashboard } from "../../hooks/useDashboard";

const DashboardHeader = () => {
  const {
    components: { ExternalLink },
  } = useDashboard();
  return (
    <div
      className={classNames(
        "flex w-screen px-4 py-3 items-center justify-between space-x-2 md:space-x-4 bg-dashboard-panel border-b border-divide print:hidden"
      )}
    >
      <SteampipeLogo />
      <div className="flex flex-grow items-center space-x-2 md:space-x-4">
        <DashboardSearch />
        <DashboardTagGroupSelect />
        <SaveSnapshotButton />
        <OpenSnapshotButton />
      </div>
      <div className="space-x-2 sm:space-x-4 md:space-x-8 flex items-center justify-end">
        <ExternalLink
          className="text-base text-foreground-lighter hover:text-foreground"
          to="https://hub.steampipe.io"
          withReferrer={true}
        >
          <>Hub</>
        </ExternalLink>
        <ExternalLink
          className="text-base text-foreground-lighter hover:text-foreground"
          to="https://steampipe.io/docs"
          withReferrer={true}
        >
          <>Docs</>
        </ExternalLink>
        <ThemeToggle />
      </div>
    </div>
  );
};

export default DashboardHeader;
